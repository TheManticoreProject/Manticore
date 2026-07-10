package llmnr

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/TheManticoreProject/Manticore/network/llmnr/class"
	"github.com/TheManticoreProject/Manticore/network/llmnr/constants"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
)

// pendingQuery holds the state the read loop needs to validate and deliver a
// response for a single outstanding query. Alongside the delivery channel it
// records the question that was asked (name/type/class) so that an incoming
// response can be checked against it: matching on the 16-bit transaction ID
// alone is insufficient (the ID space is small and the query is multicast, so
// responses are trivially forgeable and could cross-deliver a wrong answer to a
// waiting query). See RFC 4795 §2.7.
type pendingQuery struct {
	// responseChan receives the validated response for this query, together with
	// the source address it arrived from. The source is needed so that, when a
	// UDP response has its TC (truncation) bit set, the query can be retried over
	// TCP to that responder's unicast address (RFC 4795 §2.4).
	responseChan chan *udpResponse

	// name, qtype and qclass are the question that was sent. A response is only
	// delivered if its question section echoes this question. The class is
	// stored with the QU (Unicast-Preferred) bit cleared so the comparison is
	// robust: responders (notably Windows) do not necessarily echo that bit.
	name   string
	qtype  llmnr_type.Type
	qclass class.Class
}

// udpResponse pairs a validated UDP response with the address it was received
// from. The read loop populates it; Query consumes it, using the source address
// to target the TCP retry when the response is truncated.
type udpResponse struct {
	msg  *message.Message
	from *net.UDPAddr
}

// Client represents an LLMNR client that can send queries and receive responses.
//
// The Client struct provides methods to create a new client, send queries, and close the client connection.
// It manages a UDP connection and uses a sync.Map to keep track of ongoing queries.
//
// Fields:
//   - conn: A pointer to the UDP connection used for sending and receiving LLMNR messages.
//   - timeout: The duration to wait for a response before timing out.
//   - queries: A sync.Map that maps query IDs to channels for receiving responses.
//   - closeOnce: Ensures the client is closed only once.
//   - closed: A channel that is closed when the client is closed.
//   - dest: The destination address queries are sent to. It defaults to the LLMNR
//     IPv4 multicast group (224.0.0.252:5355) so that normal usage is unchanged;
//     it is overridable to allow queries to be directed at a specific responder.
//
// Usage example:
//
//	client, err := NewClient()
//	if err != nil {
//	    log.Fatalf("Failed to create client: %v", err)
//	}
//	defer client.Close()
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	resp, err := client.Query(ctx, "example.local", TypeA)
//	if err != nil {
//	    log.Fatalf("Query failed: %v", err)
//	}
//	fmt.Printf("Received response: %v\n", resp)
type Client struct {
	Conn      *net.UDPConn
	Timeout   time.Duration
	Queries   sync.Map
	CloseOnce sync.Once
	Closed    chan struct{}

	// dest is the destination address queries are written to. It defaults to
	// the LLMNR IPv4 multicast group and is overridable so queries can be
	// directed at a specific responder (e.g. for testing or unicast probing).
	dest *net.UDPAddr

	// localNets holds the IP networks configured on the host's interfaces at
	// the time the client was created. The read loop uses them to decide
	// whether a response's source address is a plausible LLMNR responder: LLMNR
	// is a link-local protocol, so a legitimate response must come from an
	// on-link (same-subnet) unicast address, a link-local address, or loopback
	// (the last covers unit tests that use a 127.0.0.1 responder).
	localNets []*net.IPNet

	// jitterInterval bounds the randomized delay applied before the first
	// transmission of a query. RFC 4795 §2.7 requires the transmission of each
	// LLMNR query to be delayed by a time randomly selected from the interval 0
	// to JITTER_INTERVAL. It defaults to constants.JitterInterval and is
	// overridable so unit tests can drive it with short, deterministic values.
	jitterInterval time.Duration

	// tcpPort is the TCP port a truncated response is retried on. RFC 4795 §2.4
	// fixes the LLMNR TCP port at 5355, so it defaults to constants.ListenPort;
	// it is overridable so unit tests can point the retry at a loopback TCP
	// responder bound to an ephemeral port.
	tcpPort int

	// retransmitInterval is the period on which an unanswered query is
	// retransmitted. RFC 4795 §2.7: "If an LLMNR query sent over UDP is not
	// resolved within LLMNR_TIMEOUT, then a sender SHOULD repeat the
	// transmission of the query in order to ensure that it was received by a
	// host capable of responding to it." It therefore defaults to the
	// LLMNR_TIMEOUT constant (constants.LLMNRTimeout) and is overridable for
	// tests.
	retransmitInterval time.Duration
}

// NewClient creates a new LLMNR client with a UDP connection.
//
// The function initializes a UDP connection for the client to use for sending and receiving LLMNR messages.
// It sets a default timeout duration for queries and starts a read loop to handle incoming responses.
//
// Returns:
//   - A pointer to the newly created Client.
//   - An error if the UDP connection could not be created.
//
// Usage example:
//
//	client, err := NewClient()
//	if err != nil {
//	    log.Fatalf("Failed to create client: %v", err)
//	}
//	defer client.Close()
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	resp, err := client.Query(ctx, "example.local", TypeA)
//	if err != nil {
//	    log.Fatalf("Query failed: %v", err)
//	}
//	fmt.Printf("Received response: %v\n", resp)
func NewClient() (*Client, error) {
	return newClient("udp4", "")
}

// NewClientForInterface creates a new LLMNR client bound to a specific address
// family and (optionally) a specific network interface.
//
// It is the entry point for querying over IPv6 multicast, which the default
// IPv4-only NewClient cannot reach. family selects the socket family and
// multicast group: "udp4" sends to the IPv4 group 224.0.0.252 (matching
// NewClient) and "udp6" sends to the IPv6 link-local group FF02::1:3. ifaceName,
// when non-empty, names the interface the queries are sent out of.
//
// An interface is effectively required for "udp6": FF02::1:3 is a link-local
// (scope 2) multicast address, so the datagram must carry a zone identifying
// the link it is sent on. On a multi-homed host the interface also overrides the
// kernel's default multicast egress interface, which is otherwise not
// necessarily the link carrying the LLMNR traffic of interest. For "udp4" the
// interface is optional and, when supplied, selects the outgoing multicast
// interface.
//
// The returned client exposes exactly the same Query and resolver API as
// NewClient; a client created with family "udp6" resolves over IPv6 (e.g.
// ResolveAAAA leaves over the IPv6 group), while responses are still validated
// as coming from a plausible on-link/link-local/loopback source.
//
// Usage example:
//
//	client, err := NewClientForInterface("udp6", "eth0")
//	if err != nil {
//	    log.Fatalf("Failed to create client: %v", err)
//	}
//	defer client.Close()
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	ips, err := client.ResolveAAAA(ctx, "host")
//	if err != nil {
//	    log.Fatalf("Query failed: %v", err)
//	}
//	fmt.Printf("Resolved: %v\n", ips)
func NewClientForInterface(family, ifaceName string) (*Client, error) {
	return newClient(family, ifaceName)
}

// newClient is the shared constructor behind NewClient and
// NewClientForInterface. It binds a UDP socket of the requested family, computes
// the multicast destination (scoping the link-local IPv6 group with the
// interface zone), selects the outgoing multicast interface when one was
// requested, and starts the read loop.
func newClient(family, ifaceName string) (*Client, error) {
	switch family {
	case "udp4", "udp6":
	default:
		return nil, fmt.Errorf("unsupported address family %q: want \"udp4\" or \"udp6\"", family)
	}

	conn, err := net.ListenUDP(family, &net.UDPAddr{})
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %w", err)
	}

	// Resolve the requested interface (if any) so it can both scope the
	// link-local multicast destination and be selected as the outgoing
	// multicast interface below.
	var iface *net.Interface
	if ifaceName != "" {
		iface, err = net.InterfaceByName(ifaceName)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to look up interface %q: %w", ifaceName, err)
		}
	}

	dest := multicastDestination(family, ifaceName)

	// Pin the outgoing multicast interface when one was requested. The
	// destination Zone already scopes a link-local IPv6 send, but selecting the
	// interface explicitly is belt-and-braces for IPv6 and the only lever for
	// IPv4 on a multi-homed host.
	if iface != nil {
		if err := setOutgoingMulticastInterface(conn, family, iface); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to select multicast interface %q: %w", ifaceName, err)
		}
	}

	c := &Client{
		Conn: conn,
		// Overall budget for a query. Per RFC 4795 §7 the LLMNR timeout is a
		// protocol constant rather than a user-tunable value, so it defaults to
		// constants.LLMNRTimeout instead of an arbitrary hardcoded duration.
		Timeout:            constants.LLMNRTimeout,
		Closed:             make(chan struct{}),
		dest:               dest,
		localNets:          localUnicastNetworks(),
		jitterInterval:     constants.JitterInterval,
		retransmitInterval: constants.LLMNRTimeout,
		tcpPort:            constants.ListenPort,
	}

	go c.readLoop()

	return c, nil
}

// multicastDestination returns the default LLMNR multicast destination for the
// given address family. For "udp4" it is the IPv4 group 224.0.0.252; for "udp6"
// it is the IPv6 link-local group FF02::1:3 with its Zone set to ifaceName.
// FF02::1:3 has link-local (scope 2) reach, so a zone identifying the link is
// required for the kernel to pick an outgoing interface for the send; leaving it
// empty yields an unscoped address that a link-local multicast send cannot use.
func multicastDestination(family, ifaceName string) *net.UDPAddr {
	if family == "udp6" {
		return &net.UDPAddr{
			IP:   net.ParseIP(constants.IPv6MulticastAddr),
			Port: constants.ListenPort,
			Zone: ifaceName,
		}
	}
	return &net.UDPAddr{
		IP:   net.ParseIP(constants.IPv4MulticastAddr),
		Port: constants.ListenPort,
	}
}

// localUnicastNetworks returns the unicast IP networks configured on the host's
// interfaces. It is best-effort: on error it returns nil, in which case the
// on-link source check simply does not match and validation falls back to the
// loopback and link-local checks.
func localUnicastNetworks() []*net.IPNet {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}

	nets := make([]*net.IPNet, 0, len(addrs))
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok {
			nets = append(nets, ipNet)
		}
	}
	return nets
}

// Close closes the client connection
func (c *Client) Close() error {
	c.CloseOnce.Do(func() {
		close(c.Closed)
		c.Conn.Close()
	})
	return nil
}

// Query sends an LLMNR query and waits for a response
func (c *Client) Query(ctx context.Context, name string, qtype llmnr_type.Type) (*message.Message, error) {
	msg := message.NewMessage()
	msg.SetQuery()
	if err := msg.AddQuestion(name, llmnr_type.Type(qtype), class.ClassIN); err != nil {
		return nil, fmt.Errorf("failed to add question: %w", err)
	}

	// Create the response channel and record the outstanding query so the read
	// loop can validate an incoming response against the exact question that was
	// asked before delivering it (the transaction ID alone is not enough).
	responseChan := make(chan *udpResponse, 1)
	pq := &pendingQuery{
		responseChan: responseChan,
		name:         name,
		qtype:        qtype,
		qclass:       class.ClassIN.BaseClass(),
	}
	c.Queries.Store(msg.Header.Identifier, pq)
	defer c.Queries.Delete(msg.Header.Identifier)

	// Encode the query once; the same bytes are (re)transmitted to the
	// configured destination (the LLMNR multicast group by default, or an
	// overridden responder address).
	encoded, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to encode message: %w", err)
	}

	// Overall budget for this query. The deadline is honored in addition to the
	// caller's context, so whichever fires first ends the wait.
	timeout := time.NewTimer(c.Timeout)
	defer timeout.Stop()

	// RFC 4795 §2.7: "the transmission of each LLMNR query and response SHOULD
	// be delayed by a time randomly selected from the interval 0 to
	// JITTER_INTERVAL." Apply that randomized delay before the first send.
	if err := c.jitterDelay(ctx, timeout); err != nil {
		return nil, err
	}

	if err := c.transmit(encoded); err != nil {
		return nil, err
	}

	// RFC 4795 §2.7: "If an LLMNR query sent over UDP is not resolved within
	// LLMNR_TIMEOUT, then a sender SHOULD repeat the transmission of the query
	// in order to ensure that it was received by a host capable of responding
	// to it." Retransmit on the LLMNR_TIMEOUT schedule until a valid response
	// arrives, the overall budget elapses, or the caller's context fires.
	ticker := time.NewTicker(c.retransmitInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout.C:
			return nil, fmt.Errorf("query timeout")
		case resp := <-responseChan:
			// RFC 4795 §2.4: "If the 'TC' bit is set in an LLMNR response, then
			// the sender SHOULD resend the LLMNR query over TCP using the unicast
			// address of the responder as the destination address." Retry over
			// TCP and, on success, deliver the complete answer in place of the
			// truncated one (discarding the truncated UDP response). If the TCP
			// retry fails, fall back to returning the truncated response rather
			// than failing the query outright, so the records that did fit are
			// still available to the caller.
			if resp.msg.Header.Flags.IsTruncation() {
				if full, err := c.queryTCP(ctx, resp.from, encoded, msg.Header.Identifier, pq); err == nil {
					return full, nil
				}
			}
			return resp.msg, nil
		case <-ticker.C:
			if err := c.transmit(encoded); err != nil {
				return nil, err
			}
		}
	}
}

// transmit writes the encoded query to the configured destination.
func (c *Client) transmit(encoded []byte) error {
	if _, err := c.Conn.WriteToUDP(encoded, c.dest); err != nil {
		return fmt.Errorf("failed to send query: %w", err)
	}
	return nil
}

// queryTCP re-issues a query over TCP after a truncated UDP response, per RFC
// 4795 §2.4. It dials the responder's unicast address (from) on the LLMNR TCP
// port, writes the same encoded query framed with the DNS-over-TCP two-byte
// length prefix (RFC 1035 §4.2.2), reads the length-prefixed full response, and
// validates it exactly as the UDP read loop does before returning it: it must be
// a response, echo the query's transaction ID, and answer the question that was
// asked. The caller's context and the client timeout bound the dial and the I/O.
//
// Parameters:
//   - ctx: the caller's context; its deadline (if any) bounds the TCP exchange.
//   - from: the unicast source address of the truncated UDP response.
//   - encoded: the exact query bytes that were sent over UDP.
//   - identifier: the query's transaction ID, used to reject a mismatched reply.
//   - pq: the pending query, used to confirm the response echoes the question.
//
// Returns the validated full response, or an error if the dial, framing, read,
// decode, or validation fails.
func (c *Client) queryTCP(ctx context.Context, from *net.UDPAddr, encoded []byte, identifier uint16, pq *pendingQuery) (*message.Message, error) {
	if from == nil || from.IP == nil {
		return nil, fmt.Errorf("no responder address for TCP retry")
	}

	port := c.tcpPort
	if port == 0 {
		port = constants.ListenPort
	}
	raddr := &net.TCPAddr{IP: from.IP, Port: port, Zone: from.Zone}

	dialer := net.Dialer{Timeout: c.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", raddr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to dial responder over TCP: %w", err)
	}
	defer conn.Close()

	// Bound the TCP exchange by the context deadline when present, otherwise by
	// the client timeout, so a stalled responder cannot hang the query.
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(c.Timeout))
	}

	if err := message.WriteTCPMessage(conn, encoded); err != nil {
		return nil, fmt.Errorf("failed to send TCP query: %w", err)
	}

	payload, err := message.ReadTCPMessage(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read TCP response: %w", err)
	}

	resp := &message.Message{}
	if _, err := resp.Unmarshal(payload); err != nil {
		return nil, fmt.Errorf("failed to decode TCP response: %w", err)
	}

	// Apply the same validation as the UDP path: the reply must be a response,
	// carry the transaction ID we asked with, and echo our question.
	if !resp.IsResponse() {
		return nil, fmt.Errorf("TCP reply is not a response")
	}
	if resp.Header.Identifier != identifier {
		return nil, fmt.Errorf("TCP response transaction ID mismatch")
	}
	if !pq.matchesQuestion(resp) {
		return nil, fmt.Errorf("TCP response does not answer the query")
	}

	return resp, nil
}

// jitterDelay blocks for a random duration in the half-open interval
// [0, jitterInterval) before returning, implementing the transmission jitter
// mandated by RFC 4795 §2.7. It returns early (with the corresponding error) if
// the caller's context is cancelled, the overall query budget elapses, or the
// client is closed, so it never delays a send past a deadline the caller cares
// about. A non-positive jitterInterval disables the delay entirely (used by
// tests that want a deterministic, immediate first send).
func (c *Client) jitterDelay(ctx context.Context, timeout *time.Timer) error {
	if c.jitterInterval <= 0 {
		return nil
	}

	d := time.Duration(rand.Int63n(int64(c.jitterInterval)))
	if d == 0 {
		return nil
	}

	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timeout.C:
		return fmt.Errorf("query timeout")
	case <-c.Closed:
		return fmt.Errorf("client closed")
	case <-t.C:
		return nil
	}
}

func (c *Client) readLoop() {
	buffer := make([]byte, constants.MaxPacketSize)
	for {
		select {
		case <-c.Closed:
			return
		default:
			n, src, err := c.Conn.ReadFromUDP(buffer)
			if err != nil {
				continue
			}

			// Drop datagrams whose source address is not a plausible LLMNR
			// responder before spending any effort parsing them.
			if !c.isPlausibleResponder(src) {
				continue
			}

			msg := message.Message{}
			_, err = msg.Unmarshal(buffer[:n])
			if err != nil {
				continue
			}

			if !msg.IsResponse() {
				continue
			}

			// Find the matching outstanding query by transaction ID, then
			// confirm the response's question section echoes the question we
			// asked before delivering it. This rejects responses that merely
			// guessed (or collided on) the 16-bit ID.
			ch, ok := c.Queries.Load(msg.Header.Identifier)
			if !ok {
				continue
			}
			pq := ch.(*pendingQuery)
			if !pq.matchesQuestion(&msg) {
				continue
			}

			// Copy the source address: ReadFromUDP hands back an address whose
			// backing storage it may reuse on the next read, and Query keeps the
			// address around to target a TCP retry when the response is truncated.
			from := &net.UDPAddr{IP: append(net.IP(nil), src.IP...), Port: src.Port, Zone: src.Zone}

			delivered := msg
			select {
			case pq.responseChan <- &udpResponse{msg: &delivered, from: from}:
			default:
			}
		}
	}
}

// isPlausibleResponder reports whether src is an acceptable source for an LLMNR
// response. LLMNR is a link-local protocol, so a legitimate unicast response
// must originate from an on-link (same-subnet) address, a link-local address,
// or loopback (which covers the loopback responder used by the unit tests). A
// multicast or unspecified source is never a valid responder.
func (c *Client) isPlausibleResponder(src *net.UDPAddr) bool {
	if src == nil || src.IP == nil {
		return false
	}
	ip := src.IP

	if ip.IsUnspecified() || ip.IsMulticast() {
		return false
	}

	// Loopback covers the in-process test responder (127.0.0.1); link-local
	// addresses (169.254.0.0/16, fe80::/10) are on the local link by definition.
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return true
	}

	// Otherwise the source must be on-link, i.e. fall inside one of the
	// networks configured on this host's interfaces (e.g. a responder on the
	// same /24 as us). This accepts genuine LAN responders while rejecting
	// off-link/routed sources.
	for _, n := range c.localNets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// matchesQuestion reports whether the response resp answers the question this
// query asked. The name is compared case-insensitively (DNS names are
// case-insensitive) and the type must match exactly. The class is compared with
// the QU (Unicast-Preferred) bit cleared, because responders do not necessarily
// echo that bit. A response with no matching question is rejected.
func (pq *pendingQuery) matchesQuestion(resp *message.Message) bool {
	for _, q := range resp.Questions {
		if !strings.EqualFold(string(q.Name), pq.name) {
			continue
		}
		if q.Type != pq.qtype {
			continue
		}
		if q.Class.BaseClass() != pq.qclass {
			continue
		}
		return true
	}
	return false
}

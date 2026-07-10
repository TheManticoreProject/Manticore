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
	// responseChan receives the validated response for this query.
	responseChan chan *message.Message

	// name, qtype and qclass are the question that was sent. A response is only
	// delivered if its question section echoes this question. The class is
	// stored with the QU (Unicast-Preferred) bit cleared so the comparison is
	// robust: responders (notably Windows) do not necessarily echo that bit.
	name   string
	qtype  llmnr_type.Type
	qclass class.Class
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
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{})
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %w", err)
	}

	c := &Client{
		Conn: conn,
		// Overall budget for a query. Per RFC 4795 §7 the LLMNR timeout is a
		// protocol constant rather than a user-tunable value, so it defaults to
		// constants.LLMNRTimeout instead of an arbitrary hardcoded duration.
		Timeout: constants.LLMNRTimeout,
		Closed:  make(chan struct{}),
		dest: &net.UDPAddr{
			IP:   net.ParseIP(constants.IPv4MulticastAddr),
			Port: constants.ListenPort,
		},
		localNets:          localUnicastNetworks(),
		jitterInterval:     constants.JitterInterval,
		retransmitInterval: constants.LLMNRTimeout,
	}

	go c.readLoop()

	return c, nil
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
	responseChan := make(chan *message.Message, 1)
	c.Queries.Store(msg.Header.Identifier, &pendingQuery{
		responseChan: responseChan,
		name:         name,
		qtype:        qtype,
		qclass:       class.ClassIN.BaseClass(),
	})
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
			return resp, nil
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

			select {
			case pq.responseChan <- &msg:
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

package nbns

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

// NetBIOS name-resolution client defaults and well-known transport parameters
// (RFC 1002 4.2.12 / RFC 1001 15.1.1).
const (
	// LimitedBroadcastAddr is the IPv4 limited-broadcast address a B-node NAME
	// QUERY REQUEST is sent to when no NBNS/WINS server is configured. Combined
	// with DefaultNBNSPort it forms the default destination 255.255.255.255:137.
	LimitedBroadcastAddr = "255.255.255.255"

	// DefaultClientTimeout bounds how long the client waits for a response to a
	// single transmission before retransmitting (RFC 1001 15.1.1 models this as
	// BCAST_REQ_RETRY_TIMEOUT / UCAST_REQ_RETRY_TIMEOUT). One second is a
	// forgiving value that works for both broadcast and unicast/WINS queries.
	DefaultClientTimeout = 1 * time.Second

	// DefaultRetransmitCount is how many times an unanswered query is
	// retransmitted after the first transmission, mirroring the RFC 1001 15.1.1
	// retry counters (BCAST_REQ_RETRY_COUNT / UCAST_REQ_RETRY_COUNT).
	DefaultRetransmitCount = 2
)

// Well-known NetBIOS name suffixes (the 16th byte of a NetBIOS name). The suffix
// selects the service registered under a name; the same base name is registered
// several times with different suffixes. These cover the suffixes an outbound
// resolver most commonly asks for.
const (
	SuffixWorkstation       byte = 0x00 // Workstation Service (the computer name)
	SuffixMessenger         byte = 0x03 // Messenger Service
	SuffixServer            byte = 0x20 // Server Service (file/print sharing)
	SuffixDomainMasterBrows byte = 0x1B // Domain Master Browser
	SuffixDomainControllers byte = 0x1C // Domain Controllers (group)
	SuffixMasterBrowser     byte = 0x1D // Master Browser
	SuffixBrowserElections  byte = 0x1E // Browser Service Elections (group)
)

// Client resolves NetBIOS names to their owner addresses by emitting NAME QUERY
// REQUEST packets (RFC 1002 4.2.12) and decoding the positive/negative NAME
// QUERY RESPONSE (RFC 1002 4.2.13). It is the outbound counterpart to the
// server-side name table in this package and mirrors the resolver shape of the
// sibling network/llmnr client (a name in, a slice of net.IP out).
//
// A Client is a lightweight, stateless value: each Resolve call opens its own
// short-lived UDP socket, so a single Client is safe to reuse across sequential
// resolutions.
type Client struct {
	// Server is the NBNS/WINS server the query is sent to (P-node/H-node style
	// unicast). It may be a bare host ("10.0.0.1") or host:port; a missing port
	// defaults to DefaultNBNSPort (137). When Server is empty the client falls
	// back to a B-node broadcast to LimitedBroadcastAddr:137.
	Server string

	// Timeout bounds the wait for a response to each individual transmission.
	// A zero value is treated as DefaultClientTimeout.
	Timeout time.Duration

	// Retransmit is the number of times an unanswered query is retransmitted
	// after the first send. A negative value is treated as zero (a single
	// transmission).
	Retransmit int

	// RecursionDesired requests, for a unicast query, that the destination
	// NBNS/WINS server perform recursive resolution (the RD bit, RFC 1002
	// 4.2.12). It is meaningful only when querying an actual name server: a
	// plain end node is not a name server and will not answer a unicast query
	// with RD set, so this defaults to false and a broadcast query always sets
	// RD regardless (the RFC B-node NAME QUERY REQUEST sets RD+B).
	RecursionDesired bool
}

// NewClient returns a Client that resolves names by B-node broadcast to
// 255.255.255.255:137, using the default timeout and retransmit count. This is
// the analogue of llmnr.NewClient for NetBIOS name resolution.
func NewClient() *Client {
	return &Client{
		Timeout:    DefaultClientTimeout,
		Retransmit: DefaultRetransmitCount,
	}
}

// NewClientWithServer returns a Client that unicasts its queries to the given
// NBNS/WINS server (P-node style) instead of broadcasting. server may be a bare
// host or host:port; a missing port defaults to 137.
func NewClientWithServer(server string) *Client {
	return &Client{
		Server:     server,
		Timeout:    DefaultClientTimeout,
		Retransmit: DefaultRetransmitCount,
	}
}

// isBroadcast reports whether this client resolves by limited broadcast (no
// server configured) rather than by unicast to a name server.
func (c *Client) isBroadcast() bool {
	return c.Server == ""
}

// encodeNameWithSuffix builds the 16-byte NetBIOS name whose first 15 bytes are
// the space-padded base name and whose 16th byte is the service suffix. This is
// the on-the-wire layout the first-level codec (NetBIOSName.FirstLevelEncode)
// expects: it copies exactly 16 bytes, so placing the suffix at index 15 here
// makes it the 16th encoded octet without the codec's space padding disturbing
// it.
func encodeNameWithSuffix(name string, suffix byte) (string, error) {
	if len(name) > NetBIOSNameLength-1 {
		return "", fmt.Errorf("name too long: %d bytes (max %d before the suffix)", len(name), NetBIOSNameLength-1)
	}

	padded := make([]byte, NetBIOSNameLength)
	copy(padded, name)
	for i := len(name); i < NetBIOSNameLength-1; i++ {
		padded[i] = ' '
	}
	padded[NetBIOSNameLength-1] = suffix

	return string(padded), nil
}

// BuildNameQueryRequest assembles a NAME QUERY REQUEST (RFC 1002 4.2.12) for the
// given base name, 16th-byte suffix, and optional scope. The packet carries a
// freshly generated transaction ID, a single NB (0x0020) question in class IN,
// and the header flags appropriate to the client's transport: OPCODE query with
// the RD (recursion desired) bit set, plus the B (broadcast) bit when the client
// is broadcasting rather than unicasting to a name server.
func (c *Client) BuildNameQueryRequest(name string, suffix byte, scope string) (*NBNSPacket, error) {
	encoded, err := encodeNameWithSuffix(name, suffix)
	if err != nil {
		return nil, err
	}

	// OPCODE is query (0x0000). A broadcast (B-node) NAME QUERY REQUEST sets the
	// B bit together with RD (RFC 1002 4.2.12). RD (recursion desired) only
	// makes sense to an actual NBNS/WINS server, and a plain end node will not
	// answer a unicast query that has RD set, so on the unicast path RD is left
	// clear unless the caller explicitly opts in via RecursionDesired.
	flags := OpNameQuery
	if c.isBroadcast() {
		flags |= FlagBroadcast | FlagRecursion
	} else if c.RecursionDesired {
		flags |= FlagRecursion
	}

	req := &NBNSPacket{
		Header: NBNSHeader{
			TransactionID: uint16(rand.Intn(0x10000)),
			Flags:         flags,
			Questions:     1,
		},
		Questions: []NBNSQuestion{
			{
				Name:  &NetBIOSName{Name: encoded, ScopeID: scope},
				Type:  QuestionTypeNB,
				Class: QuestionClassIn,
			},
		},
	}

	return req, nil
}

// destination returns the UDP address a query is sent to and whether that
// address requires the socket to be broadcast-enabled. For a unicast client it
// resolves the configured Server (defaulting the port to 137); otherwise it
// targets the IPv4 limited-broadcast address on port 137.
func (c *Client) destination() (*net.UDPAddr, bool, error) {
	if c.isBroadcast() {
		return &net.UDPAddr{IP: net.ParseIP(LimitedBroadcastAddr), Port: DefaultNBNSPort}, true, nil
	}

	// Accept both a bare host and a host:port; default the port to 137 when the
	// caller supplied only a host.
	server := c.Server
	if _, _, err := net.SplitHostPort(server); err != nil {
		server = net.JoinHostPort(server, fmt.Sprintf("%d", DefaultNBNSPort))
	}

	addr, err := net.ResolveUDPAddr("udp4", server)
	if err != nil {
		return nil, false, fmt.Errorf("failed to resolve NBNS server address %q: %w", c.Server, err)
	}
	return addr, false, nil
}

// timeout returns the effective per-transmission timeout, substituting the
// default for a non-positive configured value.
func (c *Client) timeout() time.Duration {
	if c.Timeout <= 0 {
		return DefaultClientTimeout
	}
	return c.Timeout
}

// retransmits returns the effective retransmit count, clamping a negative
// configured value to zero (a single transmission).
func (c *Client) retransmits() int {
	if c.Retransmit < 0 {
		return 0
	}
	return c.Retransmit
}

// Resolve resolves a NetBIOS name+suffix to its owner IPv4 address(es) using an
// empty scope. It is the high-level entry point mirroring llmnr.Client.Resolve:
// a name in, a slice of net.IP out.
//
// A NAME_ERROR (RCODE 0x03) negative response means the name is not registered
// with the queried responder; this is reported as a not-found result (an empty,
// non-nil slice and a nil error) rather than an error, matching the LLMNR
// resolver's no-answer semantics. A genuine failure to obtain any response
// (every transmission timing out) is returned as an error.
func (c *Client) Resolve(name string, suffix byte) ([]net.IP, error) {
	return c.ResolveWithScope(name, suffix, "")
}

// ResolveWithScope is the scope-aware form of Resolve: it resolves name+suffix
// within the given NetBIOS scope ID (pass "" for the default empty scope). See
// Resolve for the return and not-found semantics.
func (c *Client) ResolveWithScope(name string, suffix byte, scope string) ([]net.IP, error) {
	req, err := c.BuildNameQueryRequest(name, suffix, scope)
	if err != nil {
		return nil, err
	}

	data, err := req.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal NAME QUERY REQUEST: %w", err)
	}

	dest, needBroadcast, err := c.destination()
	if err != nil {
		return nil, err
	}
	if dest.IP == nil {
		return nil, fmt.Errorf("could not determine a destination address")
	}

	// A short-lived socket bound to an ephemeral local port; ReadFromUDP then
	// lets us learn each responder's source address (useful for a broadcast
	// query, where several hosts may reply).
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{})
	if err != nil {
		return nil, fmt.Errorf("failed to open UDP socket: %w", err)
	}
	defer conn.Close()

	// A limited-broadcast send requires the SO_BROADCAST socket option; enabling
	// it is a no-op on platforms where the option is not wired up.
	if needBroadcast {
		if err := enableBroadcast(conn); err != nil {
			return nil, fmt.Errorf("failed to enable broadcast on UDP socket: %w", err)
		}
	}

	// Transmit once, then retransmit up to Retransmit further times if no valid
	// response is seen within the timeout for each transmission (RFC 1001
	// 15.1.1 retry model).
	buf := make([]byte, MaxUDPSize)
	attempts := c.retransmits() + 1
	for attempt := 0; attempt < attempts; attempt++ {
		if _, err := conn.WriteToUDP(data, dest); err != nil {
			return nil, fmt.Errorf("failed to send NAME QUERY REQUEST: %w", err)
		}

		deadline := time.Now().Add(c.timeout())
		if err := conn.SetReadDeadline(deadline); err != nil {
			return nil, fmt.Errorf("failed to set read deadline: %w", err)
		}

		// Drain responses until this transmission's deadline. Datagrams that do
		// not echo our transaction ID or are not responses are ignored (a
		// broadcast query in particular may draw unrelated traffic).
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					break // deadline reached: fall through to the next retransmission
				}
				return nil, fmt.Errorf("failed to read NAME QUERY RESPONSE: %w", err)
			}

			ips, matched, err := parseNameQueryResponse(buf[:n], req.Header.TransactionID)
			if err != nil || !matched {
				continue // not ours (or undecodable): keep waiting for a valid reply
			}
			return ips, nil
		}
	}

	return nil, fmt.Errorf("no response to NAME QUERY REQUEST for %q (suffix 0x%02x)", name, suffix)
}

// parseNameQueryResponse decodes a NAME QUERY RESPONSE datagram and, if it is a
// response to our query, returns the owner IP addresses it carries. matched is
// true only when the datagram is a response echoing wantTRN; a datagram that is
// a request, or that carries a different transaction ID, yields matched=false so
// the caller keeps waiting.
//
// A negative response with RCODE NAME_ERROR (0x03) is a valid match with no
// owners: it returns an empty (non-nil) slice, matched=true, and a nil error,
// so the caller reports "not found" rather than an error. Any other non-zero
// RCODE is surfaced as an error.
func parseNameQueryResponse(data []byte, wantTRN uint16) ([]net.IP, bool, error) {
	var resp NBNSPacket
	if _, err := resp.Unmarshal(data); err != nil {
		return nil, false, err
	}

	if resp.Header.Flags&FlagResponse == 0 {
		return nil, false, nil // a request (e.g. our own broadcast echoed back)
	}
	if resp.Header.TransactionID != wantTRN {
		return nil, false, nil // response to some other query
	}

	// NAME_ERROR is the expected negative answer for an unregistered name: a
	// definitive not-found, not a failure.
	switch resp.Header.Flags & RcodeMask {
	case RcodeSuccess:
	case RcodeNameError:
		return []net.IP{}, true, nil
	default:
		return nil, true, fmt.Errorf("NAME QUERY RESPONSE returned RCODE 0x%x", resp.Header.Flags&RcodeMask)
	}

	// Collect every owner address from the NB answer records. A positive
	// response for a group name carries several ADDR_ENTRY structures in a
	// single record's RDATA, so walk the RDATA in 6-byte ADDR_ENTRY strides
	// rather than decoding only the first entry.
	ips := make([]net.IP, 0)
	for _, rr := range resp.Answers {
		if rr.Type != QuestionTypeNB {
			continue
		}
		for off := 0; off+6 <= len(rr.RData); off += 6 {
			ip, err := ParseIPFromRData(rr.RData[off : off+6])
			if err != nil {
				continue
			}
			ips = append(ips, ip)
		}
	}

	return ips, true, nil
}

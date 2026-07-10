package llmnr

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/llmnr/constants"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message/header"
)

// startTCPResponder binds a loopback TCP socket that acts as a fake LLMNR TCP
// responder. For each accepted connection it reads one DNS-over-TCP framed query
// (RFC 1035 §4.2.2), passes the query bytes to respond, and writes whatever
// respond returns (if non-nil) back framed with the two-byte length prefix. It
// returns the listener's TCP port (to be injected as the client's tcpPort) and a
// cleanup function that stops the accept loop and closes the socket.
func startTCPResponder(t *testing.T, respond func(query []byte) []byte) (int, func()) {
	t.Helper()

	ln, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("failed to start TCP responder: %v", err)
	}

	done := make(chan struct{})
	go func() {
		for {
			conn, err := ln.Accept()
			select {
			case <-done:
				return
			default:
			}
			if err != nil {
				continue
			}
			go func(c net.Conn) {
				defer c.Close()
				query, err := message.ReadTCPMessage(c)
				if err != nil {
					return
				}
				if reply := respond(query); reply != nil {
					_ = message.WriteTCPMessage(c, reply)
				}
			}(conn)
		}
	}()

	return ln.Addr().(*net.TCPAddr).Port, func() {
		close(done)
		ln.Close()
	}
}

// truncatedAnswer builds a well-formed LLMNR response that echoes the query's
// transaction ID and question but carries the TC (truncation) bit set and no
// answer records, simulating a responder whose full answer did not fit in a UDP
// datagram. Returns nil if the query cannot be parsed or has no question.
func truncatedAnswer(query []byte) []byte {
	m := message.NewMessage()
	if _, err := m.Unmarshal(query); err != nil || len(m.Questions) == 0 {
		return nil
	}

	resp := message.NewMessage()
	resp.Header.Identifier = m.Header.Identifier
	resp.SetResponse()
	resp.Header.Flags |= header.FlagTC
	// Echo the question so the client's read loop accepts the response as
	// matching the outstanding query before acting on the TC bit.
	if err := resp.AddQuestion(string(m.Questions[0].Name), m.Questions[0].Type, m.Questions[0].Class); err != nil {
		return nil
	}

	encoded, err := resp.Marshal()
	if err != nil {
		return nil
	}
	return encoded
}

// TestClientTCPFallbackOnTruncation drives the RFC 4795 §2.4 TCP fallback end to
// end over loopback: a fake UDP responder answers with the TC bit set and no
// records, and a fake TCP responder serves the complete answer. The client must
// notice the TC bit, retry the same query over TCP to the responder's address,
// and deliver the full untruncated answer in place of the truncated UDP one.
func TestClientTCPFallbackOnTruncation(t *testing.T) {
	// The TCP responder serves the complete answer (owner name + A record).
	tcpPort, tcpCleanup := startTCPResponder(t, func(query []byte) []byte {
		return echoAnswer(query, "10.7.0.10")
	})
	defer tcpCleanup()

	// The UDP responder answers with the TC bit set, triggering the TCP retry.
	udpAddr, udpCleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		return truncatedAnswer(query)
	})
	defer udpCleanup()

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()
	c.dest = udpAddr
	c.jitterInterval = 0
	// Point the TCP retry at the loopback TCP responder's ephemeral port instead
	// of the fixed RFC port 5355.
	c.tcpPort = tcpPort

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := c.Query(ctx, "host.local", llmnr_type.TypeA)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	// The delivered response must be the full TCP answer, not the truncated UDP
	// one: it carries an answer record and does not have the TC bit set.
	if resp.Header.Flags.IsTruncation() {
		t.Error("Query() returned the truncated UDP response, want the full TCP answer")
	}
	if len(resp.Answers) != 1 {
		t.Fatalf("Query() answers = %d, want 1 (from the TCP responder)", len(resp.Answers))
	}
	if resp.Answers[0].Name != "host.local" {
		t.Errorf("Query() answer name = %q, want %q", resp.Answers[0].Name, "host.local")
	}
	if got := net.IP(resp.Answers[0].RData).String(); got != "10.7.0.10" {
		t.Errorf("Query() answer RDATA = %q, want %q", got, "10.7.0.10")
	}
}

// TestClientTCPFallbackFailureReturnsTruncated confirms that when the TC bit is
// set but the TCP retry cannot be completed (nothing is listening on the TCP
// port), Query degrades gracefully by returning the truncated UDP response
// rather than failing outright.
func TestClientTCPFallbackFailureReturnsTruncated(t *testing.T) {
	udpAddr, udpCleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		return truncatedAnswer(query)
	})
	defer udpCleanup()

	// Reserve a TCP port and immediately release it so the dial is refused.
	ln, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("failed to reserve TCP port: %v", err)
	}
	deadPort := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()
	c.dest = udpAddr
	c.jitterInterval = 0
	c.tcpPort = deadPort
	c.Timeout = 500 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := c.Query(ctx, "host.local", llmnr_type.TypeA)
	if err != nil {
		t.Fatalf("Query() error = %v, want the truncated response returned on TCP failure", err)
	}
	if !resp.Header.Flags.IsTruncation() {
		t.Error("Query() response missing TC bit, want the truncated UDP response on TCP failure")
	}
}

// startResponder binds a loopback UDP socket that acts as a fake LLMNR responder
// and invokes respond for each datagram it receives. Whatever respond returns (if
// non-nil) is sent back to the datagram's source, so tests never touch the network
// or the real multicast group. It returns the responder's address (to be injected
// as the client's destination) and a cleanup function that stops the goroutine and
// closes the socket.
func startResponder(t *testing.T, respond func(query []byte, from *net.UDPAddr) []byte) (*net.UDPAddr, func()) {
	t.Helper()

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("failed to start responder: %v", err)
	}

	done := make(chan struct{})
	go func() {
		buf := make([]byte, constants.MaxPacketSize)
		for {
			// A short read deadline lets the loop notice the done signal even
			// when no datagram arrives (e.g. the timeout tests).
			_ = conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
			n, addr, err := conn.ReadFromUDP(buf)
			select {
			case <-done:
				return
			default:
			}
			if err != nil {
				continue
			}
			if reply := respond(append([]byte(nil), buf[:n]...), addr); reply != nil {
				_, _ = conn.WriteToUDP(reply, addr)
			}
		}
	}()

	return conn.LocalAddr().(*net.UDPAddr), func() {
		close(done)
		conn.Close()
	}
}

// echoAnswer builds a well-formed LLMNR response for the given query, echoing the
// transaction identifier and answering the first question with a Type A record
// pointing at ip. Returns nil if the query cannot be parsed or carries no question.
func echoAnswer(query []byte, ip string) []byte {
	m := message.NewMessage()
	if _, err := m.Unmarshal(query); err != nil {
		return nil
	}
	if len(m.Questions) == 0 {
		return nil
	}

	resp := message.NewMessage()
	resp.Header.Identifier = m.Header.Identifier
	resp.SetResponse()
	if err := resp.AddAnswerClassINTypeA(string(m.Questions[0].Name), ip); err != nil {
		return nil
	}

	encoded, err := resp.Marshal()
	if err != nil {
		return nil
	}
	return encoded
}

// TestNewClient verifies that a freshly created client is wired with a live UDP
// connection, the default timeout, a non-nil close channel, and a destination that
// defaults to the LLMNR IPv4 multicast group on the RFC 4795 port.
func TestNewClient(t *testing.T) {
	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()

	if c.Conn == nil {
		t.Error("NewClient() Conn is nil")
	}
	if c.Timeout != constants.LLMNRTimeout {
		t.Errorf("NewClient() Timeout = %v, want %v", c.Timeout, constants.LLMNRTimeout)
	}
	if c.Closed == nil {
		t.Error("NewClient() Closed channel is nil")
	}
	if c.dest == nil {
		t.Fatal("NewClient() dest is nil")
	}
	if got := c.dest.IP.String(); got != constants.IPv4MulticastAddr {
		t.Errorf("NewClient() dest IP = %q, want %q", got, constants.IPv4MulticastAddr)
	}
	if c.dest.Port != constants.ListenPort {
		t.Errorf("NewClient() dest Port = %d, want %d", c.dest.Port, constants.ListenPort)
	}
}

// startResponder6 is the IPv6 analogue of startResponder: it binds a udp6
// loopback socket (::1) acting as a fake LLMNR responder and replies to each
// datagram with whatever respond returns. It exercises the client's IPv6 send
// path without touching the real FF02::1:3 multicast group.
func startResponder6(t *testing.T, respond func(query []byte, from *net.UDPAddr) []byte) (*net.UDPAddr, func()) {
	t.Helper()

	conn, err := net.ListenUDP("udp6", &net.UDPAddr{IP: net.IPv6loopback})
	if err != nil {
		t.Fatalf("failed to start IPv6 responder: %v", err)
	}

	done := make(chan struct{})
	go func() {
		buf := make([]byte, constants.MaxPacketSize)
		for {
			_ = conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
			n, addr, err := conn.ReadFromUDP(buf)
			select {
			case <-done:
				return
			default:
			}
			if err != nil {
				continue
			}
			if reply := respond(append([]byte(nil), buf[:n]...), addr); reply != nil {
				_, _ = conn.WriteToUDP(reply, addr)
			}
		}
	}()

	return conn.LocalAddr().(*net.UDPAddr), func() {
		close(done)
		conn.Close()
	}
}

// echoAnswerAAAA builds a well-formed LLMNR response echoing the query's
// transaction ID and answering the first question with a Type AAAA record
// pointing at ip. Returns nil if the query cannot be parsed or has no question.
func echoAnswerAAAA(query []byte, ip string) []byte {
	m := message.NewMessage()
	if _, err := m.Unmarshal(query); err != nil {
		return nil
	}
	if len(m.Questions) == 0 {
		return nil
	}

	resp := message.NewMessage()
	resp.Header.Identifier = m.Header.Identifier
	resp.SetResponse()
	if err := resp.AddAnswerClassINTypeAAAA(string(m.Questions[0].Name), ip); err != nil {
		return nil
	}

	encoded, err := resp.Marshal()
	if err != nil {
		return nil
	}
	return encoded
}

// TestMulticastDestination checks the per-family default destination: IPv4 uses
// the 224.0.0.252 group with no zone, while IPv6 uses the FF02::1:3 link-local
// group with the interface name appended as the zone (scope), which is required
// for a link-local multicast send. The zone must surface in the address's string
// form as "%iface".
func TestMulticastDestination(t *testing.T) {
	v4 := multicastDestination("udp4", "")
	if got := v4.IP.String(); got != constants.IPv4MulticastAddr {
		t.Errorf("multicastDestination(udp4) IP = %q, want %q", got, constants.IPv4MulticastAddr)
	}
	if v4.Port != constants.ListenPort {
		t.Errorf("multicastDestination(udp4) Port = %d, want %d", v4.Port, constants.ListenPort)
	}
	if v4.Zone != "" {
		t.Errorf("multicastDestination(udp4) Zone = %q, want empty", v4.Zone)
	}

	v6 := multicastDestination("udp6", "eth0")
	if !v6.IP.Equal(net.ParseIP(constants.IPv6MulticastAddr)) {
		t.Errorf("multicastDestination(udp6) IP = %q, want %q", v6.IP, constants.IPv6MulticastAddr)
	}
	if v6.Port != constants.ListenPort {
		t.Errorf("multicastDestination(udp6) Port = %d, want %d", v6.Port, constants.ListenPort)
	}
	if v6.Zone != "eth0" {
		t.Errorf("multicastDestination(udp6) Zone = %q, want %q", v6.Zone, "eth0")
	}
	// The zone must appear in the wire/string form so the kernel scopes the send
	// to the named link (e.g. "[ff02::1:3%eth0]:5355").
	if got := v6.String(); !strings.Contains(got, "%eth0") {
		t.Errorf("multicastDestination(udp6) String = %q, want it to contain %q", got, "%eth0")
	}
}

// TestNewClientForInterfaceRejectsBadFamily confirms an unsupported address
// family is rejected rather than silently binding the wrong socket.
func TestNewClientForInterfaceRejectsBadFamily(t *testing.T) {
	if _, err := NewClientForInterface("udp7", ""); err == nil {
		t.Error("NewClientForInterface(udp7) error = nil, want an unsupported-family error")
	}
}

// TestNewClientForInterfaceIPv6 constructs an IPv6 client bound to the loopback
// interface (present on every host, so the test stays offline) and verifies it
// targets the FF02::1:3 group on the RFC 4795 port with the interface set as the
// destination zone. This also exercises selecting the outgoing multicast
// interface, which must not fail construction.
func TestNewClientForInterfaceIPv6(t *testing.T) {
	c, err := NewClientForInterface("udp6", "lo")
	if err != nil {
		t.Fatalf("NewClientForInterface(udp6, lo) error = %v", err)
	}
	defer c.Close()

	if c.Conn == nil {
		t.Error("NewClientForInterface() Conn is nil")
	}
	if c.dest == nil {
		t.Fatal("NewClientForInterface() dest is nil")
	}
	if !c.dest.IP.Equal(net.ParseIP(constants.IPv6MulticastAddr)) {
		t.Errorf("NewClientForInterface() dest IP = %q, want %q", c.dest.IP, constants.IPv6MulticastAddr)
	}
	if c.dest.Port != constants.ListenPort {
		t.Errorf("NewClientForInterface() dest Port = %d, want %d", c.dest.Port, constants.ListenPort)
	}
	if c.dest.Zone != "lo" {
		t.Errorf("NewClientForInterface() dest Zone = %q, want %q", c.dest.Zone, "lo")
	}
}

// TestNewClientForInterfaceUnknownInterface confirms that naming an interface
// that does not exist fails construction (and does not leak the socket).
func TestNewClientForInterfaceUnknownInterface(t *testing.T) {
	if _, err := NewClientForInterface("udp6", "definitely-not-an-iface0"); err == nil {
		t.Error("NewClientForInterface() error = nil, want an interface-lookup error")
	}
}

// TestClientQueryDispatchIPv6 drives a full query through an IPv6 (udp6) client
// against a ::1 loopback responder, proving the IPv6 send/receive path resolves
// end to end. The responder answers with an AAAA record and the client's source
// validation must accept the loopback responder.
func TestClientQueryDispatchIPv6(t *testing.T) {
	addr, cleanup := startResponder6(t, func(query []byte, _ *net.UDPAddr) []byte {
		return echoAnswerAAAA(query, "fe80::1")
	})
	defer cleanup()

	c, err := NewClientForInterface("udp6", "")
	if err != nil {
		t.Fatalf("NewClientForInterface(udp6) error = %v", err)
	}
	defer c.Close()
	// Direct the query at the loopback responder instead of the FF02::1:3 group.
	c.dest = addr
	c.jitterInterval = 0

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := c.Query(ctx, "host", llmnr_type.TypeAAAA)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(resp.Answers) != 1 {
		t.Fatalf("Query() answers = %d, want 1", len(resp.Answers))
	}
	if got := net.IP(resp.Answers[0].RData).String(); got != "fe80::1" {
		t.Errorf("Query() answer RDATA = %q, want %q", got, "fe80::1")
	}
}

// TestResolveAAAAOverIPv6 exercises the resolver helper over an IPv6 client and
// the ::1 loopback responder, confirming ResolveAAAA leaves over the udp6 socket
// and decodes the AAAA answer.
func TestResolveAAAAOverIPv6(t *testing.T) {
	addr, cleanup := startResponder6(t, respondWithAnswers(nil, []string{"fe80::1234"}))
	defer cleanup()

	c, err := NewClientForInterface("udp6", "")
	if err != nil {
		t.Fatalf("NewClientForInterface(udp6) error = %v", err)
	}
	defer c.Close()
	c.dest = addr
	c.jitterInterval = 0

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ips, err := c.ResolveAAAA(ctx, "host")
	if err != nil {
		t.Fatalf("ResolveAAAA() error = %v", err)
	}
	if len(ips) != 1 || !containsIP(ips, "fe80::1234") {
		t.Errorf("ResolveAAAA() = %v, want [fe80::1234]", ipsToStrings(ips))
	}
}

// TestClientQueryDispatch drives a full query against a loopback responder and
// checks that the matching response is dispatched back through Query, including
// the answered owner name and A record.
func TestClientQueryDispatch(t *testing.T) {
	addr, cleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		return echoAnswer(query, "10.7.0.10")
	})
	defer cleanup()

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()
	c.dest = addr

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := c.Query(ctx, "host.local", llmnr_type.TypeA)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if !resp.IsResponse() {
		t.Error("Query() returned a message that is not a response")
	}
	if len(resp.Answers) != 1 {
		t.Fatalf("Query() answers = %d, want 1", len(resp.Answers))
	}
	if resp.Answers[0].Name != "host.local" {
		t.Errorf("Query() answer name = %q, want %q", resp.Answers[0].Name, "host.local")
	}
	if got := net.IP(resp.Answers[0].RData).String(); got != "10.7.0.10" {
		t.Errorf("Query() answer RDATA = %q, want %q", got, "10.7.0.10")
	}
}

// TestClientQueryMatchesRealCapture feeds a genuine captured Windows Server 2016
// LLMNR response through the client. The responder replays the exact captured
// bytes (only rewriting the transaction ID so it matches the client's randomly
// generated query ID) so this is a known-answer test of the client dispatch path
// against real traffic: owner name TMP-W-2016, A record 10.7.0.10.
func TestClientQueryMatchesRealCapture(t *testing.T) {
	// Real LLMNR response captured from TMP-W-2016 (see message_test.go).
	const liveResponseHex = "37c8800000010001000000000a544d502d572d3230313600000100010a544d502d572d3230313600000100010000001e00040a07000a"
	capture, err := hex.DecodeString(liveResponseHex)
	if err != nil {
		t.Fatalf("failed to decode capture fixture: %v", err)
	}

	addr, cleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		if len(query) < 2 {
			return nil
		}
		reply := append([]byte(nil), capture...)
		// Rewrite the response transaction ID to match the incoming query so it
		// dispatches to the waiting Query call.
		reply[0], reply[1] = query[0], query[1]
		return reply
	})
	defer cleanup()

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()
	c.dest = addr

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := c.Query(ctx, "TMP-W-2016", llmnr_type.TypeA)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(resp.Answers) != 1 {
		t.Fatalf("Query() answers = %d, want 1", len(resp.Answers))
	}
	if resp.Answers[0].Name != "TMP-W-2016" {
		t.Errorf("Query() answer name = %q, want %q", resp.Answers[0].Name, "TMP-W-2016")
	}
	if got := net.IP(resp.Answers[0].RData).String(); got != "10.7.0.10" {
		t.Errorf("Query() answer RDATA = %q, want %q", got, "10.7.0.10")
	}
}

// TestClientConcurrentQueries fires many queries in parallel through one client
// and checks that each caller receives the response bearing its own transaction
// ID, exercising the sync.Map dispatch under concurrency. The responder answers
// each query with a distinct IP derived from the queried name.
func TestClientConcurrentQueries(t *testing.T) {
	addr, cleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		m := message.NewMessage()
		if _, err := m.Unmarshal(query); err != nil || len(m.Questions) == 0 {
			return nil
		}
		// The name is "host-N.local"; answer with 10.0.0.N so the caller can
		// confirm it received the response matching its own query.
		name := string(m.Questions[0].Name)
		var idx int
		if _, err := fmt.Sscanf(name, "host-%d.local", &idx); err != nil {
			return nil
		}
		return echoAnswer(query, fmt.Sprintf("10.0.0.%d", idx))
	})
	defer cleanup()

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()
	c.dest = addr

	const n = 16
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 1; i <= n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			resp, err := c.Query(ctx, fmt.Sprintf("host-%d.local", i), llmnr_type.TypeA)
			if err != nil {
				errs <- fmt.Errorf("query %d: %w", i, err)
				return
			}
			if len(resp.Answers) != 1 {
				errs <- fmt.Errorf("query %d: got %d answers, want 1", i, len(resp.Answers))
				return
			}
			want := fmt.Sprintf("10.0.0.%d", i)
			if got := net.IP(resp.Answers[0].RData).String(); got != want {
				errs <- fmt.Errorf("query %d: answer RDATA = %q, want %q", i, got, want)
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

// TestClientQueryTimeout confirms Query returns a timeout error when the responder
// never replies. The responder returns nil (no datagram) and the client timeout is
// shortened so the test stays fast.
func TestClientQueryTimeout(t *testing.T) {
	addr, cleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		return nil // silent responder
	})
	defer cleanup()

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()
	c.dest = addr
	c.Timeout = 150 * time.Millisecond

	_, err = c.Query(context.Background(), "host.local", llmnr_type.TypeA)
	if err == nil {
		t.Fatal("Query() error = nil, want timeout")
	}
	if err.Error() != "query timeout" {
		t.Errorf("Query() error = %q, want %q", err.Error(), "query timeout")
	}
}

// TestClientQueryContextCancel confirms Query honours a cancelled context and
// returns the context error rather than waiting for the timeout.
func TestClientQueryContextCancel(t *testing.T) {
	addr, cleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		return nil // silent responder
	})
	defer cleanup()

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()
	c.dest = addr
	c.Timeout = 5 * time.Second // long, so the context cancellation is what fires

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before the call

	_, err = c.Query(ctx, "host.local", llmnr_type.TypeA)
	if err != context.Canceled {
		t.Errorf("Query() error = %v, want %v", err, context.Canceled)
	}
}

// TestClientReadLoopFiltersNonResponse verifies that the read loop discards LLMNR
// messages that are not responses (QR bit clear) even when the transaction ID
// matches an outstanding query, so a stray query on the wire cannot satisfy a
// pending Query. The responder echoes the query ID but leaves QR clear, so Query
// must time out.
func TestClientReadLoopFiltersNonResponse(t *testing.T) {
	addr, cleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		m := message.NewMessage()
		if _, err := m.Unmarshal(query); err != nil {
			return nil
		}
		// Build a reply that keeps the same ID but stays a query (QR clear).
		reply := message.NewMessage()
		reply.Header.Identifier = m.Header.Identifier
		reply.SetQuery()
		if err := reply.AddAnswerClassINTypeA("host.local", "10.7.0.10"); err != nil {
			return nil
		}
		encoded, err := reply.Marshal()
		if err != nil {
			return nil
		}
		return encoded
	})
	defer cleanup()

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()
	c.dest = addr
	c.Timeout = 200 * time.Millisecond

	_, err = c.Query(context.Background(), "host.local", llmnr_type.TypeA)
	if err == nil || err.Error() != "query timeout" {
		t.Errorf("Query() error = %v, want %q (non-response must be filtered)", err, "query timeout")
	}
}

// TestClientReadLoopFiltersUnknownID verifies that a response whose transaction ID
// does not match any outstanding query is dropped and cannot satisfy a pending
// Query. The responder answers with a fixed, non-matching ID, so Query must time
// out.
func TestClientReadLoopFiltersUnknownID(t *testing.T) {
	addr, cleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		reply := echoAnswer(query, "10.7.0.10")
		if reply == nil {
			return nil
		}
		// Overwrite the transaction ID with a value the client is not waiting on.
		reply[0], reply[1] = 0xDE, 0xAD
		return reply
	})
	defer cleanup()

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()
	c.dest = addr
	c.Timeout = 200 * time.Millisecond

	_, err = c.Query(context.Background(), "host.local", llmnr_type.TypeA)
	if err == nil || err.Error() != "query timeout" {
		t.Errorf("Query() error = %v, want %q (unknown ID must be dropped)", err, "query timeout")
	}
}

// TestClientReadLoopFiltersMismatchedQuestion verifies that a response bearing a
// matching transaction ID but a question section that does not correspond to the
// outstanding query is rejected, so an attacker who guesses (or collides on) the
// 16-bit ID cannot inject an answer for a name the client never asked about. The
// responder echoes the query ID but answers a different owner name, so Query must
// time out.
func TestClientReadLoopFiltersMismatchedQuestion(t *testing.T) {
	addr, cleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		m := message.NewMessage()
		if _, err := m.Unmarshal(query); err != nil {
			return nil
		}
		// Same transaction ID, but answer a name the client did not query.
		resp := message.NewMessage()
		resp.Header.Identifier = m.Header.Identifier
		resp.SetResponse()
		if err := resp.AddAnswerClassINTypeA("attacker.local", "10.7.0.66"); err != nil {
			return nil
		}
		encoded, err := resp.Marshal()
		if err != nil {
			return nil
		}
		return encoded
	})
	defer cleanup()

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()
	c.dest = addr
	c.Timeout = 200 * time.Millisecond

	_, err = c.Query(context.Background(), "host.local", llmnr_type.TypeA)
	if err == nil || err.Error() != "query timeout" {
		t.Errorf("Query() error = %v, want %q (mismatched question must be rejected)", err, "query timeout")
	}
}

// TestClientReadLoopFiltersMismatchedType verifies that a response whose question
// echoes the queried name but with a different record type is rejected, so a Type
// AAAA answer cannot satisfy a pending Type A query sharing the same ID.
func TestClientReadLoopFiltersMismatchedType(t *testing.T) {
	addr, cleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		m := message.NewMessage()
		if _, err := m.Unmarshal(query); err != nil {
			return nil
		}
		// Same ID and name, but answer with a Type AAAA record instead of A.
		resp := message.NewMessage()
		resp.Header.Identifier = m.Header.Identifier
		resp.SetResponse()
		if err := resp.AddAnswerClassINTypeAAAA("host.local", "fe80::1"); err != nil {
			return nil
		}
		encoded, err := resp.Marshal()
		if err != nil {
			return nil
		}
		return encoded
	})
	defer cleanup()

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()
	c.dest = addr
	c.Timeout = 200 * time.Millisecond

	_, err = c.Query(context.Background(), "host.local", llmnr_type.TypeA)
	if err == nil || err.Error() != "query timeout" {
		t.Errorf("Query() error = %v, want %q (mismatched question type must be rejected)", err, "query timeout")
	}
}

// TestClientReadLoopMatchesQuestionCaseInsensitively confirms a genuine response
// is still delivered when the echoed question name differs only in letter case
// from the query, since DNS names are case-insensitive. A responder that would
// upper-case the name (as some implementations do) must not be rejected.
func TestClientReadLoopMatchesQuestionCaseInsensitively(t *testing.T) {
	addr, cleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		m := message.NewMessage()
		if _, err := m.Unmarshal(query); err != nil || len(m.Questions) == 0 {
			return nil
		}
		// Answer using an upper-cased owner name to exercise case-insensitive
		// matching against the lower-case query name.
		resp := message.NewMessage()
		resp.Header.Identifier = m.Header.Identifier
		resp.SetResponse()
		if err := resp.AddAnswerClassINTypeA(strings.ToUpper(string(m.Questions[0].Name)), "10.7.0.10"); err != nil {
			return nil
		}
		encoded, err := resp.Marshal()
		if err != nil {
			return nil
		}
		return encoded
	})
	defer cleanup()

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()
	c.dest = addr

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := c.Query(ctx, "host.local", llmnr_type.TypeA)
	if err != nil {
		t.Fatalf("Query() error = %v (case-insensitive question must match)", err)
	}
	if len(resp.Answers) != 1 {
		t.Fatalf("Query() answers = %d, want 1", len(resp.Answers))
	}
	if got := net.IP(resp.Answers[0].RData).String(); got != "10.7.0.10" {
		t.Errorf("Query() answer RDATA = %q, want %q", got, "10.7.0.10")
	}
}

// TestClientIsPlausibleResponder exercises the source-address sanity check
// directly: loopback and link-local sources are accepted, multicast/unspecified
// sources are rejected, and an off-link routed source is rejected while an
// on-link source (inside a configured interface network) is accepted.
func TestClientIsPlausibleResponder(t *testing.T) {
	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()

	// Constrain localNets to a known network so the on-link case is deterministic.
	_, onLink, err := net.ParseCIDR("10.7.0.0/24")
	if err != nil {
		t.Fatalf("ParseCIDR() error = %v", err)
	}
	c.localNets = []*net.IPNet{onLink}

	cases := []struct {
		name string
		ip   net.IP
		want bool
	}{
		{"loopback", net.IPv4(127, 0, 0, 1), true},
		{"link-local", net.IPv4(169, 254, 1, 2), true},
		{"on-link unicast", net.IPv4(10, 7, 0, 10), true},
		{"off-link unicast", net.IPv4(8, 8, 8, 8), false},
		{"multicast", net.ParseIP(constants.IPv4MulticastAddr), false},
		{"unspecified", net.IPv4zero, false},
		{"nil", nil, false},
	}
	for _, tc := range cases {
		var src *net.UDPAddr
		if tc.ip != nil {
			src = &net.UDPAddr{IP: tc.ip, Port: constants.ListenPort}
		}
		if got := c.isPlausibleResponder(src); got != tc.want {
			t.Errorf("isPlausibleResponder(%s) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestClientCloseIdempotent confirms Close can be called multiple times without
// panicking (double close of the channel) and always returns nil.
func TestClientCloseIdempotent(t *testing.T) {
	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if err := c.Close(); err != nil {
		t.Errorf("first Close() = %v, want nil", err)
	}
	if err := c.Close(); err != nil {
		t.Errorf("second Close() = %v, want nil", err)
	}

	// The Closed channel must be closed after Close().
	select {
	case <-c.Closed:
	default:
		t.Error("Closed channel is not closed after Close()")
	}
}

// TestClientQueryRetransmits confirms that, per RFC 4795 §2.7, an unanswered
// query is retransmitted more than once before the overall budget elapses. A
// silent responder counts every datagram it receives; with short, overridden
// jitter/retransmit values the client must send the query multiple times within
// the timeout window. This exercises the retransmission schedule deterministically
// and fast, by COUNTING the datagrams that reach the loopback responder.
func TestClientQueryRetransmits(t *testing.T) {
	var count int64
	addr, cleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		atomic.AddInt64(&count, 1)
		return nil // never answer, so the client keeps retransmitting
	})
	defer cleanup()

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()
	c.dest = addr

	// Short, deterministic timing so the test is fast and not flaky: a tiny
	// jitter, a 20ms retransmit interval, and a 130ms overall budget yield
	// roughly seven transmissions (t≈0, 20, 40, 60, 80, 100, 120ms).
	c.jitterInterval = 1 * time.Millisecond
	c.retransmitInterval = 20 * time.Millisecond
	c.Timeout = 130 * time.Millisecond

	_, err = c.Query(context.Background(), "host.local", llmnr_type.TypeA)
	if err == nil || err.Error() != "query timeout" {
		t.Fatalf("Query() error = %v, want %q", err, "query timeout")
	}

	if got := atomic.LoadInt64(&count); got < 2 {
		t.Errorf("responder received %d datagrams, want >1 (query must be retransmitted)", got)
	}
}

// TestClientQueryJitterBound confirms the first transmission is delayed but still
// occurs within the JitterInterval bound mandated by RFC 4795 §2.7 (a delay
// randomly selected from [0, JitterInterval)). The responder records how long
// after the Query call the first datagram arrives; with a large retransmit
// interval only the initial send lands inside the observation window, so the
// measured delay is the jitter delay itself and must not exceed JitterInterval.
func TestClientQueryJitterBound(t *testing.T) {
	firstRecv := make(chan time.Duration, 1)
	var start atomic.Int64
	addr, cleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		if s := start.Load(); s != 0 {
			select {
			case firstRecv <- time.Duration(time.Now().UnixNano() - s):
			default:
			}
		}
		return nil // silent, so no response cuts the window short
	})
	defer cleanup()

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()
	c.dest = addr

	const jitter = 60 * time.Millisecond
	c.jitterInterval = jitter
	c.retransmitInterval = 1 * time.Second // keep the second send out of the window
	c.Timeout = 500 * time.Millisecond

	start.Store(time.Now().UnixNano())
	go func() { _, _ = c.Query(context.Background(), "host.local", llmnr_type.TypeA) }()

	select {
	case d := <-firstRecv:
		// Allow modest scheduling slack on top of the jitter bound.
		if d > jitter+50*time.Millisecond {
			t.Errorf("first transmission delayed %v, want <= %v (JitterInterval bound)", d, jitter)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("first transmission never observed at responder")
	}
}

// TestClientQueryContextDeadline confirms the caller's context deadline is honored
// during the retransmission loop: with a silent responder and a long overall
// timeout, a short context deadline must end the query promptly (returning the
// context error) rather than waiting for Timeout.
func TestClientQueryContextDeadline(t *testing.T) {
	addr, cleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		return nil // silent responder
	})
	defer cleanup()

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()
	c.dest = addr
	c.jitterInterval = 1 * time.Millisecond
	c.retransmitInterval = 20 * time.Millisecond
	c.Timeout = 5 * time.Second // long, so the context deadline is what fires

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	begin := time.Now()
	_, err = c.Query(ctx, "host.local", llmnr_type.TypeA)
	elapsed := time.Since(begin)

	if err != context.DeadlineExceeded {
		t.Errorf("Query() error = %v, want %v", err, context.DeadlineExceeded)
	}
	if elapsed > time.Second {
		t.Errorf("Query() took %v, want it to return near the 60ms context deadline, not the 5s timeout", elapsed)
	}
}

// TestClientQueryInvalidName confirms Query surfaces an error for a name that
// fails domain-name validation and never touches the network.
func TestClientQueryInvalidName(t *testing.T) {
	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer c.Close()

	// A label longer than 63 octets is invalid per the LLMNR/DNS label rules.
	tooLong := ""
	for i := 0; i < 64; i++ {
		tooLong += "a"
	}

	if _, err := c.Query(context.Background(), tooLong, llmnr_type.TypeA); err == nil {
		t.Error("Query() error = nil, want validation error for over-long label")
	}
}

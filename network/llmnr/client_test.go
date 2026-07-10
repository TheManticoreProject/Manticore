package llmnr

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/llmnr/constants"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
)

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
	if c.Timeout != 2*time.Second {
		t.Errorf("NewClient() Timeout = %v, want %v", c.Timeout, 2*time.Second)
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

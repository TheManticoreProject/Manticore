package llmnr

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
)

// newTestClient returns a client whose destination is addr and whose jitter is
// disabled, so resolver tests drive the loopback responder deterministically and
// without any pre-send delay.
func newTestClient(t *testing.T, addr *net.UDPAddr) *Client {
	t.Helper()

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	c.dest = addr
	c.jitterInterval = 0
	return c
}

// respondWithAnswers builds a responder that echoes the query's transaction ID
// and question and answers it with the supplied A and AAAA addresses, so a
// single responder can serve ResolveA, ResolveAAAA and Resolve.
func respondWithAnswers(aIPs []string, aaaaIPs []string) func(query []byte, from *net.UDPAddr) []byte {
	return func(query []byte, _ *net.UDPAddr) []byte {
		m := message.NewMessage()
		if _, err := m.Unmarshal(query); err != nil || len(m.Questions) == 0 {
			return nil
		}
		name := string(m.Questions[0].Name)

		resp := message.NewMessage()
		resp.Header.Identifier = m.Header.Identifier
		resp.SetResponse()

		// Echo the exact question that was asked so the client's response
		// validation accepts the reply for the pending query.
		if err := resp.AddQuestion(name, m.Questions[0].Type, m.Questions[0].Class); err != nil {
			return nil
		}
		for _, ip := range aIPs {
			if err := resp.AddAnswerClassINTypeA(name, ip); err != nil {
				return nil
			}
		}
		for _, ip := range aaaaIPs {
			if err := resp.AddAnswerClassINTypeAAAA(name, ip); err != nil {
				return nil
			}
		}

		encoded, err := resp.Marshal()
		if err != nil {
			return nil
		}
		return encoded
	}
}

// ipsToStrings renders a slice of net.IP as strings for assertions.
func ipsToStrings(ips []net.IP) []string {
	out := make([]string, 0, len(ips))
	for _, ip := range ips {
		out = append(out, ip.String())
	}
	return out
}

// containsIP reports whether want (an IP string) appears in got.
func containsIP(got []net.IP, want string) bool {
	for _, ip := range got {
		if ip.String() == want {
			return true
		}
	}
	return false
}

// TestResolveAReturnsIPv4 checks that ResolveA collects every A answer into the
// returned []net.IP.
func TestResolveAReturnsIPv4(t *testing.T) {
	addr, cleanup := startResponder(t, respondWithAnswers([]string{"10.7.0.10", "10.7.0.11"}, nil))
	defer cleanup()

	c := newTestClient(t, addr)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ips, err := c.ResolveA(ctx, "host.local")
	if err != nil {
		t.Fatalf("ResolveA() error = %v", err)
	}
	if len(ips) != 2 {
		t.Fatalf("ResolveA() returned %d IPs (%v), want 2", len(ips), ipsToStrings(ips))
	}
	if !containsIP(ips, "10.7.0.10") || !containsIP(ips, "10.7.0.11") {
		t.Errorf("ResolveA() = %v, want to contain 10.7.0.10 and 10.7.0.11", ipsToStrings(ips))
	}
}

// TestResolveAAAAReturnsIPv6 checks that ResolveAAAA collects AAAA answers.
func TestResolveAAAAReturnsIPv6(t *testing.T) {
	addr, cleanup := startResponder(t, respondWithAnswers(nil, []string{"fe80::1"}))
	defer cleanup()

	c := newTestClient(t, addr)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ips, err := c.ResolveAAAA(ctx, "host.local")
	if err != nil {
		t.Fatalf("ResolveAAAA() error = %v", err)
	}
	if len(ips) != 1 {
		t.Fatalf("ResolveAAAA() returned %d IPs (%v), want 1", len(ips), ipsToStrings(ips))
	}
	if !containsIP(ips, "fe80::1") {
		t.Errorf("ResolveAAAA() = %v, want to contain fe80::1", ipsToStrings(ips))
	}
}

// TestResolveReturnsBothFamilies checks that Resolve concatenates the A and AAAA
// results from the two concurrent queries.
func TestResolveReturnsBothFamilies(t *testing.T) {
	addr, cleanup := startResponder(t, respondWithAnswers([]string{"10.7.0.10"}, []string{"fe80::1"}))
	defer cleanup()

	c := newTestClient(t, addr)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ips, err := c.Resolve(ctx, "host.local")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(ips) != 2 {
		t.Fatalf("Resolve() returned %d IPs (%v), want 2", len(ips), ipsToStrings(ips))
	}
	if !containsIP(ips, "10.7.0.10") || !containsIP(ips, "fe80::1") {
		t.Errorf("Resolve() = %v, want to contain 10.7.0.10 and fe80::1", ipsToStrings(ips))
	}
}

// TestResolveAFiltersMismatchedName verifies that answers whose owner name does
// not match the queried name are filtered out, so a responder that pads the
// answer section with an unrelated A record cannot leak into the result.
func TestResolveAFiltersMismatchedName(t *testing.T) {
	addr, cleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		m := message.NewMessage()
		if _, err := m.Unmarshal(query); err != nil || len(m.Questions) == 0 {
			return nil
		}
		name := string(m.Questions[0].Name)

		resp := message.NewMessage()
		resp.Header.Identifier = m.Header.Identifier
		resp.SetResponse()
		if err := resp.AddQuestion(name, m.Questions[0].Type, m.Questions[0].Class); err != nil {
			return nil
		}
		// The requested name resolves to 10.7.0.10; an unrelated name is also
		// present and must be discarded by the resolver's name filter.
		if err := resp.AddAnswerClassINTypeA(name, "10.7.0.10"); err != nil {
			return nil
		}
		if err := resp.AddAnswerClassINTypeA("other.local", "10.7.0.99"); err != nil {
			return nil
		}
		encoded, err := resp.Marshal()
		if err != nil {
			return nil
		}
		return encoded
	})
	defer cleanup()

	c := newTestClient(t, addr)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ips, err := c.ResolveA(ctx, "host.local")
	if err != nil {
		t.Fatalf("ResolveA() error = %v", err)
	}
	if len(ips) != 1 || !containsIP(ips, "10.7.0.10") {
		t.Errorf("ResolveA() = %v, want exactly [10.7.0.10]", ipsToStrings(ips))
	}
}

// TestResolveANoAnswerReturnsEmpty verifies the documented no-answer semantics:
// a response that echoes the question but carries no matching record yields an
// empty, non-nil slice and a nil error rather than an error.
func TestResolveANoAnswerReturnsEmpty(t *testing.T) {
	addr, cleanup := startResponder(t, func(query []byte, _ *net.UDPAddr) []byte {
		m := message.NewMessage()
		if _, err := m.Unmarshal(query); err != nil || len(m.Questions) == 0 {
			return nil
		}
		resp := message.NewMessage()
		resp.Header.Identifier = m.Header.Identifier
		resp.SetResponse()
		// Echo the question but supply no answers at all.
		if err := resp.AddQuestion(string(m.Questions[0].Name), m.Questions[0].Type, m.Questions[0].Class); err != nil {
			return nil
		}
		encoded, err := resp.Marshal()
		if err != nil {
			return nil
		}
		return encoded
	})
	defer cleanup()

	c := newTestClient(t, addr)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ips, err := c.ResolveA(ctx, "host.local")
	if err != nil {
		t.Fatalf("ResolveA() error = %v, want nil for a no-answer response", err)
	}
	if ips == nil {
		t.Fatal("ResolveA() returned a nil slice, want empty non-nil slice")
	}
	if len(ips) != 0 {
		t.Errorf("ResolveA() = %v, want an empty slice", ipsToStrings(ips))
	}
}

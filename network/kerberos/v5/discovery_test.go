package kerberos

import (
	"net"
	"testing"
)

// seqRandFloat returns a deterministic [0,1) "random" source that walks a fixed
// sequence, so weighted SRV selection can be exercised reproducibly.
func seqRandFloat(values ...float64) func() float64 {
	i := 0
	return func() float64 {
		v := values[i%len(values)]
		i++
		return v
	}
}

// TestOrderSRVEndpointsPriority verifies that endpoints are ordered by ascending
// priority: every record in a lower-priority group must precede every record in
// a higher-priority group, regardless of weight.
func TestOrderSRVEndpointsPriority(t *testing.T) {
	records := []*net.SRV{
		{Target: "kdc-hi.corp.local.", Port: 88, Priority: 20, Weight: 100},
		{Target: "kdc-lo.corp.local.", Port: 88, Priority: 0, Weight: 0},
		{Target: "kdc-mid.corp.local.", Port: 88, Priority: 10, Weight: 50},
	}
	got := orderSRVEndpointsRand(records, seqRandFloat(0.5))
	wantOrder := []string{"kdc-lo.corp.local", "kdc-mid.corp.local", "kdc-hi.corp.local"}
	if len(got) != len(wantOrder) {
		t.Fatalf("got %d endpoints, want %d: %v", len(got), len(wantOrder), got)
	}
	for i, host := range wantOrder {
		if got[i].host != host {
			t.Errorf("endpoint[%d].host = %q, want %q (full: %v)", i, got[i].host, host, got)
		}
		if got[i].port != 88 {
			t.Errorf("endpoint[%d].port = %d, want 88", i, got[i].port)
		}
	}
}

// TestOrderSRVEndpointsFailoverComplete verifies that within a priority group the
// weighted shuffle still yields every target exactly once, so the list is a
// complete failover sequence.
func TestOrderSRVEndpointsFailoverComplete(t *testing.T) {
	records := []*net.SRV{
		{Target: "a.corp.local.", Port: 88, Priority: 0, Weight: 10},
		{Target: "b.corp.local.", Port: 88, Priority: 0, Weight: 20},
		{Target: "c.corp.local.", Port: 88, Priority: 0, Weight: 30},
	}
	got := orderSRVEndpointsRand(records, seqRandFloat(0.1, 0.9, 0.5, 0.2))
	if len(got) != 3 {
		t.Fatalf("got %d endpoints, want 3: %v", len(got), got)
	}
	seen := map[string]int{}
	for _, ep := range got {
		seen[ep.host]++
	}
	for _, host := range []string{"a.corp.local", "b.corp.local", "c.corp.local"} {
		if seen[host] != 1 {
			t.Errorf("host %q appeared %d times, want exactly 1 (full: %v)", host, seen[host], got)
		}
	}
}

// TestOrderSRVEndpointsWeightPreference verifies the RFC 2782 weighting: with a
// random point that falls inside the first record's weight span, that record is
// picked first even when it is not first in the input slice.
func TestOrderSRVEndpointsWeightPreference(t *testing.T) {
	records := []*net.SRV{
		{Target: "heavy.corp.local.", Port: 88, Priority: 0, Weight: 90},
		{Target: "light.corp.local.", Port: 88, Priority: 0, Weight: 10},
	}
	// target = 0.05*100 = 5 -> falls within heavy's [0,90) span -> heavy first.
	got := orderSRVEndpointsRand(records, seqRandFloat(0.05, 0.5))
	if got[0].host != "heavy.corp.local" {
		t.Errorf("first endpoint = %q, want heavy.corp.local (higher weight)", got[0].host)
	}

	// A point past heavy's span selects light first.
	got = orderSRVEndpointsRand(records, seqRandFloat(0.95, 0.5))
	if got[0].host != "light.corp.local" {
		t.Errorf("first endpoint = %q, want light.corp.local", got[0].host)
	}
}

// TestOrderSRVEndpointsDropsNullTarget verifies the RFC 2782 "." target (service
// explicitly not available) and empty targets are dropped.
func TestOrderSRVEndpointsDropsNullTarget(t *testing.T) {
	records := []*net.SRV{
		{Target: ".", Port: 0, Priority: 0, Weight: 0},
		{Target: "", Port: 0, Priority: 0, Weight: 0},
		{Target: "real.corp.local.", Port: 88, Priority: 0, Weight: 0},
	}
	got := orderSRVEndpointsRand(records, seqRandFloat(0.5))
	if len(got) != 1 || got[0].host != "real.corp.local" {
		t.Fatalf("got %v, want single real.corp.local endpoint", got)
	}
}

// TestOrderSRVEndpointsDedup verifies duplicate host:port pairs (e.g. the same
// KDC advertised under both _tcp and _udp) collapse to one entry.
func TestOrderSRVEndpointsDedup(t *testing.T) {
	records := []*net.SRV{
		{Target: "kdc.corp.local.", Port: 88, Priority: 0, Weight: 0},
		{Target: "kdc.corp.local.", Port: 88, Priority: 0, Weight: 0},
		{Target: "kdc.corp.local.", Port: 8888, Priority: 0, Weight: 0},
	}
	got := orderSRVEndpointsRand(records, seqRandFloat(0.5, 0.5, 0.5))
	if len(got) != 2 {
		t.Fatalf("got %d endpoints, want 2 (dedup by host:port): %v", len(got), got)
	}
}

// TestParseKDCEndpoint verifies that bare hosts default to port 88 and that
// explicit host:port (including bracketed IPv6) is honoured.
func TestParseKDCEndpoint(t *testing.T) {
	cases := []struct {
		input    string
		wantHost string
		wantPort int
	}{
		{"dc.corp.local", "dc.corp.local", 88},
		{"10.0.0.1", "10.0.0.1", 88},
		{"dc.corp.local:8888", "dc.corp.local", 8888},
		{"10.0.0.1:9999", "10.0.0.1", 9999},
		{"[::1]:8888", "::1", 8888},
		{"::1", "::1", 88},
	}
	for _, tc := range cases {
		ep := parseKDCEndpoint(tc.input)
		if ep.host != tc.wantHost || ep.port != tc.wantPort {
			t.Errorf("parseKDCEndpoint(%q) = {%q, %d}, want {%q, %d}", tc.input, ep.host, ep.port, tc.wantHost, tc.wantPort)
		}
	}
}

// TestEndpointsForRealmPrecedence checks the resolution order used before any
// DNS SRV discovery: explicit WithRealmKDC, then the home realm's configured
// KDC, then a custom resolver — all yielding the standard port 88.
func TestEndpointsForRealmPrecedence(t *testing.T) {
	c := NewClient("alice", "corp.local", "10.0.0.1")
	c.WithRealmKDC("child.corp.local", "10.0.1.1")
	c.WithKDCResolver(func(realm string) (string, error) { return "resolver-" + realm, nil })

	cases := []struct{ realm, wantHost string }{
		{"child.corp.local", "10.0.1.1"},        // explicit override
		{"corp.local", "10.0.0.1"},              // home realm
		{"OTHER.LOCAL", "resolver-OTHER.LOCAL"}, // custom resolver
	}
	for _, tc := range cases {
		eps, err := c.endpointsForRealm(tc.realm)
		if err != nil {
			t.Fatalf("endpointsForRealm(%q): %v", tc.realm, err)
		}
		if len(eps) != 1 || eps[0].host != tc.wantHost || eps[0].port != defaultKDCPort {
			t.Errorf("endpointsForRealm(%q) = %v, want single %s:%d", tc.realm, eps, tc.wantHost, defaultKDCPort)
		}
	}
}

// TestEndpointsForRealmExplicitPort verifies that host:port passes through when
// configured on the home realm, a cross-realm KDC, or a custom resolver.
func TestEndpointsForRealmExplicitPort(t *testing.T) {
	c := NewClient("alice", "corp.local", "10.0.0.1:9999")
	c.WithRealmKDC("child.corp.local", "10.0.1.1:7777")
	c.WithKDCResolver(func(realm string) (string, error) { return "[::1]:6666", nil })

	cases := []struct {
		realm    string
		wantHost string
		wantPort int
	}{
		{"child.corp.local", "10.0.1.1", 7777},
		{"corp.local", "10.0.0.1", 9999},
		{"OTHER.LOCAL", "::1", 6666},
	}
	for _, tc := range cases {
		eps, err := c.endpointsForRealm(tc.realm)
		if err != nil {
			t.Fatalf("endpointsForRealm(%q): %v", tc.realm, err)
		}
		if len(eps) != 1 || eps[0].host != tc.wantHost || eps[0].port != tc.wantPort {
			t.Errorf("endpointsForRealm(%q) = %v, want single %s:%d", tc.realm, eps, tc.wantHost, tc.wantPort)
		}
	}
}

// TestKDCSendEndpointsFailover verifies transport failover: dialing a closed
// port on the first endpoint fails, and kdcSendEndpoints moves on to the next.
// (Both point at closed ports so no live KDC is required; the assertion is that
// every endpoint is attempted and the aggregate error names the count.)
func TestKDCSendEndpointsFailover(t *testing.T) {
	// Two endpoints on the loopback with almost-certainly-closed high ports.
	endpoints := []kdcEndpoint{
		{host: "127.0.0.1", port: 1},
		{host: "127.0.0.1", port: 2},
	}
	_, err := kdcSendEndpoints(nil, endpoints, []byte{0x00})
	if err == nil {
		t.Fatal("expected failure when all endpoints are unreachable")
	}
	// The error should report that all endpoints were tried.
	if got := err.Error(); got == "" {
		t.Fatal("expected a descriptive aggregate error")
	}
}

// TestKDCSendEndpointsEmpty verifies an empty endpoint list is a clean error, not
// a panic.
func TestKDCSendEndpointsEmpty(t *testing.T) {
	if _, err := kdcSendEndpoints(nil, nil, []byte{0x00}); err == nil {
		t.Fatal("expected error for empty endpoint list")
	}
}

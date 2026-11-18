package ldap

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type recordingHostnameResolver struct {
	addresses map[string][]string
	errors    map[string]error
	lookups   []string
	deadlines []bool
}

func (r *recordingHostnameResolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	r.lookups = append(r.lookups, host)
	_, hasDeadline := ctx.Deadline()
	r.deadlines = append(r.deadlines, hasDeadline)
	return r.addresses[host], r.errors[host]
}

func TestResolveDomainControllerHostnames(t *testing.T) {
	resolver := &recordingHostnameResolver{
		addresses: map[string][]string{
			"dc1.example.test": {"192.0.2.10", "2001:db8::10", "not-an-address"},
			"dc2.example.test": {"192.0.2.20", "192.0.2.10"},
		},
		errors: map[string]error{"offline.example.test": errors.New("not found")},
	}
	controllers := map[string][]string{
		"dc1": {"dc1.example.test"},
		"dc2": {"dc2.example.test", "offline.example.test"},
	}

	got := resolveDomainControllerHostnames(resolver, controllers)
	want := []string{"192.0.2.10", "192.0.2.20", "2001:db8::10"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resolveDomainControllerHostnames() = %v, want %v", got, want)
	}
	if len(resolver.lookups) != 3 {
		t.Fatalf("resolver calls = %v, want one per hostname", resolver.lookups)
	}
	for i, hasDeadline := range resolver.deadlines {
		if !hasDeadline {
			t.Fatalf("lookup %d had no timeout deadline", i)
		}
	}
}

package ldap

import "testing"

// TestKerberosSPNHostname checks that the ldap SPN is built from the connection
// host by default and from the override when one is set. The override lets a
// Kerberos bind reach a DC by IP while naming it by FQDN in the SPN, which is what
// Active Directory registers.
func TestKerberosSPNHostname(t *testing.T) {
	s, err := NewSession("192.168.1.7", 636, nil, true, true)
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	if got := s.kerberosSPNHostname(); got != "192.168.1.7" {
		t.Errorf("default SPN hostname = %q, want the connection host", got)
	}

	s.SetKerberosSPNHostname("dc1.corp.local")
	if got := s.kerberosSPNHostname(); got != "dc1.corp.local" {
		t.Errorf("SPN hostname = %q, want the override", got)
	}

	// Clearing it falls back to the connection host.
	s.SetKerberosSPNHostname("")
	if got := s.kerberosSPNHostname(); got != "192.168.1.7" {
		t.Errorf("SPN hostname after clear = %q, want the connection host", got)
	}
}

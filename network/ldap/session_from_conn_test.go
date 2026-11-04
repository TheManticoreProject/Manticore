package ldap

import (
	"testing"

	"github.com/go-ldap/ldap/v3"
)

// TestNewSessionFromConn checks the constructor wraps the given connection and
// defaults the page size, so the attack layer (which needs only the connection
// and the page size) works on an externally-authenticated conn.
func TestNewSessionFromConn(t *testing.T) {
	conn := &ldap.Conn{}

	s := NewSessionFromConn(conn)
	if s == nil {
		t.Fatal("NewSessionFromConn returned nil")
	}
	if s.connection != conn {
		t.Fatal("the session does not wrap the provided connection")
	}
	if s.GetPageSize() != DefaultPageSize {
		t.Fatalf("GetPageSize() = %d, want DefaultPageSize (%d)", s.GetPageSize(), DefaultPageSize)
	}
	// No credentials are held: the bind happened elsewhere.
	if s.credentials != nil {
		t.Fatal("the session should hold no credentials")
	}
}

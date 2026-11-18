package ldap

import (
	"net"
	"testing"
	"time"

	ber "github.com/go-asn1-ber/asn1-ber"
	ldapv3 "github.com/go-ldap/ldap/v3"
)

// TestSessionCloseSendsUnbindRequest verifies that a normal session shutdown is
// visible to the LDAP server as an UnbindRequest rather than only as a TCP close.
func TestSessionCloseSendsUnbindRequest(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %s", err)
	}
	defer listener.Close()

	protocolOp := make(chan ber.Tag, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			protocolOp <- 0
			return
		}
		defer conn.Close()

		packet, readErr := ber.ReadPacket(conn)
		if readErr != nil || len(packet.Children) < 2 {
			protocolOp <- 0
			return
		}
		protocolOp <- packet.Children[1].Tag
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial recording server: %s", err)
	}
	ldapConn := ldapv3.NewConn(conn, false)
	ldapConn.Start()

	session := NewSessionFromConn(ldapConn)
	session.Close()

	select {
	case tag := <-protocolOp:
		if tag != ldapv3.ApplicationUnbindRequest {
			t.Fatalf("protocol operation tag = %d, want UnbindRequest (%d)", tag, ldapv3.ApplicationUnbindRequest)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for LDAP shutdown request")
	}

	// Preserve the previous closed-connection behavior for callers that use or
	// close the Session again: go-ldap should return an error, not a nil-pointer
	// panic, and repeated shutdown remains harmless.
	if _, err := session.Query("DC=example,DC=com", "(objectClass=*)", []string{"cn"}, ldapv3.ScopeBaseObject); err == nil {
		t.Fatal("query on a closed session unexpectedly succeeded")
	}
	session.Close()
}

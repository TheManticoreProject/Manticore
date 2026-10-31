package ldap

import (
	"net"
	"testing"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/go-ldap/ldap/v3"
)

// recordSearchScopes runs search against a session connected to a minimal LDAP
// server and returns the scope each SearchRequest carried on the wire, in order.
func recordSearchScopes(t *testing.T, search func(session *Session)) []int64 {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %s", err)
	}
	defer listener.Close()

	var scopes []int64
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			packet, err := ber.ReadPacket(conn)
			if err != nil || len(packet.Children) < 2 {
				return
			}
			messageID, ok := packet.Children[0].Value.(int64)
			if !ok {
				return
			}
			switch packet.Children[1].Tag {
			case ldap.ApplicationSearchRequest:
				// SearchRequest ::= SEQUENCE { baseObject, scope ENUMERATED, ... }
				if len(packet.Children[1].Children) > 1 {
					if scope, ok := packet.Children[1].Children[1].Value.(int64); ok {
						scopes = append(scopes, scope)
					}
				}
				conn.Write(ldapResultDone(messageID, ldap.ApplicationSearchResultDone))
			case ldap.ApplicationUnbindRequest:
				return
			default:
				conn.Write(ldapResultDone(messageID, ldap.ApplicationBindResponse))
			}
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial the recording server: %s", err)
	}
	connection := ldap.NewConn(conn, false)
	connection.Start()

	search(&Session{connection: connection})

	connection.Close()
	<-done

	return scopes
}

// TestQueryChildrenScope checks that QueryChildren searches only the immediate
// children of the base object (singleLevel) rather than silently widening the
// search to the whole subtree.
func TestQueryChildrenScope(t *testing.T) {
	scopes := recordSearchScopes(t, func(session *Session) {
		if _, err := session.QueryChildren("DC=example,DC=com", "(objectClass=*)", []string{"cn"}); err != nil {
			t.Errorf("QueryChildren returned an error: %s", err)
		}
	})

	if len(scopes) != 1 {
		t.Fatalf("expected 1 search, got %d", len(scopes))
	}
	if scopes[0] != int64(ldap.ScopeSingleLevel) {
		t.Fatalf("QueryChildren searched with scope %d (%s), want %d (%s)",
			scopes[0], ldap.ScopeMap[int(scopes[0])], ldap.ScopeSingleLevel, ldap.ScopeMap[ldap.ScopeSingleLevel])
	}
}

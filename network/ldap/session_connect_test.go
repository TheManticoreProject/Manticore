package ldap

import (
	"net"
	"testing"
	"time"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/go-ldap/ldap/v3"

	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// bindRejecter is a minimal LDAP server that answers every BindRequest with
// invalidCredentials (49). It accepts a single connection and closes done once
// the client side of that connection is gone, which is how a test observes
// whether the client released its socket.
type bindRejecter struct {
	listener net.Listener
	done     chan struct{}
}

func newBindRejecter(t *testing.T) *bindRejecter {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %s", err)
	}

	rejecter := &bindRejecter{listener: listener, done: make(chan struct{})}
	go rejecter.serve()

	return rejecter
}

func (r *bindRejecter) serve() {
	defer close(r.done)

	conn, err := r.listener.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		packet, err := ber.ReadPacket(conn)
		if err != nil {
			return
		}
		if len(packet.Children) < 2 {
			return
		}

		messageID, ok := packet.Children[0].Value.(int64)
		if !ok {
			return
		}

		if packet.Children[1].Tag != ldap.ApplicationBindRequest {
			return
		}

		message := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Response")
		message.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, messageID, "MessageID"))
		response := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ldap.ApplicationBindResponse, nil, "Response")
		response.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, int64(ldap.LDAPResultInvalidCredentials), "resultCode"))
		response.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", "matchedDN"))
		response.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", "diagnosticMessage"))
		message.AppendChild(response)
		conn.Write(message.Bytes())
	}
}

// TestConnectClosesConnectionOnBindFailure checks that a rejected bind releases
// the TCP connection for every non-Kerberos bind method. Connect never hands the
// connection to the session on failure, so nothing else can close it.
func TestConnectClosesConnectionOnBindFailure(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		password string
		hashes   string
	}{
		{name: "pass the hash", username: "alice", hashes: ":31d6cfe0d16ae931b73c59d7e0c089c0"},
		{name: "simple bind", username: "alice", password: "wrong"},
		{name: "unauthenticated bind", username: "alice"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			rejecter := newBindRejecter(t)
			defer rejecter.listener.Close()

			creds, err := credentials.NewCredentials("LAB", testCase.username, testCase.password, testCase.hashes)
			if err != nil {
				t.Fatalf("failed to create the credentials: %s", err)
			}

			port := rejecter.listener.Addr().(*net.TCPAddr).Port
			session, err := NewSession("127.0.0.1", port, creds, false, false)
			if err != nil {
				t.Fatalf("failed to create the session: %s", err)
			}

			ok, err := session.Connect()
			if ok || err == nil {
				t.Fatalf("expected the bind to be rejected, got ok=%v err=%v", ok, err)
			}

			select {
			case <-rejecter.done:
			case <-time.After(5 * time.Second):
				t.Fatalf("the connection was left open after the failed bind: %s", err)
			}
		})
	}
}

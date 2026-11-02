package client

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/gssapi"
)

// keyedSecCtx is a fakeSecCtx that also reports a session key, as the Kerberos
// context does once its handshake is complete.
type keyedSecCtx struct {
	fakeSecCtx
	key []byte
}

func (k *keyedSecCtx) SessionKey() []byte { return k.key }

// TestBindAdoptsTheProvidersSessionKey checks a provider that knows its session
// key (Kerberos) has it reported by Client.SessionKey after the bind, which is
// what DRSUAPI decrypts replicated secrets with. Before, only NTLM set it.
func TestBindAdoptsTheProvidersSessionKey(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckBytes(t, 4280, 4280))
	c := NewClient(ft)
	key := bytes.Repeat([]byte{0x5A}, 32)
	if err := c.SetAuthProvider(pdu.AuthTypeGSSKerberos, pdu.AuthLevelPktPrivacy, &keyedSecCtx{fakeSecCtx: fakeSecCtx{tokenLen: 16}, key: key}, []byte("ap-req")); err != nil {
		t.Fatal(err)
	}
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if !bytes.Equal(c.SessionKey(), key) {
		t.Fatalf("SessionKey() = %x, want the provider's %x", c.SessionKey(), key)
	}
}

// TestKerberosSessionKeyPrefersTheSubkey checks the Kerberos context reports the
// negotiated subkey (the acceptor's, adopted from its AP-REP) and falls back to
// the ticket session key. Live DCSync over a Kerberos binding against Windows
// Server 2012 R2 and 2025 decrypts the right NT hash with the 32-byte subkey.
func TestKerberosSessionKeyPrefersTheSubkey(t *testing.T) {
	ticketKey, subKey := bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{2}, 32)
	k := &kerberosSecurityContext{ctx: &gssapi.SecContext{SessionKey: ticketKey, SubKey: subKey}}
	if !bytes.Equal(k.SessionKey(), subKey) {
		t.Errorf("SessionKey() = %x, want the subkey", k.SessionKey())
	}
	k = &kerberosSecurityContext{ctx: &gssapi.SecContext{SessionKey: ticketKey}}
	if !bytes.Equal(k.SessionKey(), ticketKey) {
		t.Errorf("SessionKey() without a subkey = %x, want the ticket session key", k.SessionKey())
	}
}

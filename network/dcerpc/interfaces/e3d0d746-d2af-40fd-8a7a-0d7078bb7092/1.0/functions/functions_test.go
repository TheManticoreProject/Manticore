package functions_test

import (
	"bytes"
	"testing"

	BitsPeerAuth "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3d0d746-d2af-40fd-8a7a-0d7078bb7092/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3d0d746-d2af-40fd-8a7a-0d7078bb7092/1.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// responder is an ndr.Invoker that records the marshalled request stub (so the on-the-wire
// NDR layout can be asserted) and replies with a canned response stub.
type responder struct {
	stub  []byte
	opnum uint16
	resp  []byte
}

func (r *responder) Invoke(in ndr.Call, out any) error {
	b, err := ndr.Request(in)
	if err != nil {
		return err
	}
	r.stub = b
	r.opnum = in.Opnum()
	if r.resp == nil {
		return nil
	}
	return ndr.Response(r.resp, out)
}

// serverKeyStub is the response wire shape of ExchangePublicKeys ([MS-BPAU] 3.2.4.1):
// the inline [out, ref] KEY_LENGTH pServerKeyLength, then the [unique] pointer to the
// conformant byte array (pServerKey), then the trailing HRESULT. The size_is count is
// derived from the slice on marshal, exactly as the server encodes it.
type serverKeyStub struct {
	PServerKeyLength uint32
	PServerKey       []byte `ndr:"unique,size_is=PServerKeyLength"`
	Status           uint32 `ndr:"retval"`
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := ndr.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	return b
}

// TestExchangePublicKeysRoundTrip drives a full request/response: the client sends its
// certificate blob and the server returns its own. It pins the opnum, checks the client
// key bytes appear in the request stub, and asserts the server key is decoded intact.
func TestExchangePublicKeysRoundTrip(t *testing.T) {
	clientKey := []byte{0xca, 0xfe, 0xba, 0xbe, 0x01, 0x02, 0x03}
	serverKey := []byte{0xde, 0xad, 0xbe, 0xef, 0x11, 0x22}

	r := &responder{resp: mustMarshal(t, &serverKeyStub{PServerKey: serverKey})}
	got, err := functions.ExchangePublicKeys(r, clientKey)
	if err != nil {
		t.Fatalf("ExchangePublicKeys: %v", err)
	}
	if r.opnum != BitsPeerAuth.OpnumExchangePublicKeys {
		t.Fatalf("opnum = %d, want %d", r.opnum, BitsPeerAuth.OpnumExchangePublicKeys)
	}
	if !bytes.Equal(got, serverKey) {
		t.Fatalf("server key = %x, want %x", got, serverKey)
	}
	// Request stub begins with ClientKeyLength (7), a non-null referent id for the
	// [unique] ClientKey, then the conformant max_count (7), then the key bytes.
	if len(r.stub) < 8 || r.stub[0] != 0x07 {
		t.Fatalf("expected ClientKeyLength=7 at start, got % x", r.stub[:min(8, len(r.stub))])
	}
	if !bytes.Contains(r.stub, clientKey) {
		t.Fatalf("client key %x not found in request stub %x", clientKey, r.stub)
	}
}

// TestExchangePublicKeysNilClientKey verifies that declining to send a certificate
// (nil clientKey) marshals ClientKeyLength=0 and a NULL [unique] referent for ClientKey.
func TestExchangePublicKeysNilClientKey(t *testing.T) {
	r := &responder{resp: mustMarshal(t, &serverKeyStub{PServerKey: []byte{0x01}})}
	if _, err := functions.ExchangePublicKeys(r, nil); err != nil {
		t.Fatalf("ExchangePublicKeys: %v", err)
	}
	// ClientKeyLength (0) then the ClientKey [unique] referent id (0 => NULL): 8 zero bytes.
	if len(r.stub) < 8 {
		t.Fatalf("request stub too short: % x", r.stub)
	}
	for i := 0; i < 8; i++ {
		if r.stub[i] != 0 {
			t.Fatalf("expected zero ClientKeyLength + NULL referent, got % x", r.stub[:8])
		}
	}
}

// TestExchangePublicKeysErrorStatus verifies a nonzero HRESULT is surfaced as an error
// and the (NULL) server key is not returned.
func TestExchangePublicKeysErrorStatus(t *testing.T) {
	r := &responder{resp: mustMarshal(t, &serverKeyStub{Status: BitsPeerAuth.StatusAccessDenied})}
	got, err := functions.ExchangePublicKeys(r, []byte{0x01})
	if err == nil {
		t.Fatalf("expected error for HRESULT 0x%08x, got nil (key=%x)", BitsPeerAuth.StatusAccessDenied, got)
	}
	if got != nil {
		t.Fatalf("server key should be nil on error, got %x", got)
	}
}

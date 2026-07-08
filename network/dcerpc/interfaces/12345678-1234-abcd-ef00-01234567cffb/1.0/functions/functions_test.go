package functions_test

import (
	"bytes"
	"testing"

	netlogon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// captureInvoker records the marshalled request stub and opnum without any network I/O, so
// the on-the-wire NDR layout of the Netlogon requests can be asserted.
type captureInvoker struct {
	stub  []byte
	opnum uint16
}

func (c *captureInvoker) Invoke(in ndr.Call, out any) error {
	b, err := ndr.Request(in)
	if err != nil {
		return err
	}
	c.stub = b
	c.opnum = in.Opnum()
	return nil
}

// TestNetrServerReqChallengeMarshal checks the opnum, the trailing fixed 8-byte client
// challenge, and that the computer name is carried as a UTF-16LE NDR string.
func TestNetrServerReqChallengeMarshal(t *testing.T) {
	cap := &captureInvoker{}
	challenge := msnrpc.NETLOGON_CREDENTIAL{1, 2, 3, 4, 5, 6, 7, 8}

	if _, _, err := functions.NetrServerReqChallenge(cap, "DC", "DC", challenge); err != nil {
		t.Fatalf("NetrServerReqChallenge: %v", err)
	}
	if cap.opnum != netlogon.OpnumNetrServerReqChallenge {
		t.Fatalf("opnum = %d, want %d", cap.opnum, netlogon.OpnumNetrServerReqChallenge)
	}
	if got := cap.stub[len(cap.stub)-8:]; !bytes.Equal(got, challenge[:]) {
		t.Fatalf("trailing challenge = %x, want %x", got, challenge[:])
	}
	if !bytes.Contains(cap.stub, []byte{'D', 0, 'C', 0}) {
		t.Fatalf("UTF-16LE %q not found in stub %x", "DC", cap.stub)
	}
}

// TestNetrServerAuthenticate2Marshal checks the enum width and that the negotiate flags and
// credential are present, exercising the enum + fixed-array + scalar field tags together.
func TestNetrServerAuthenticate2Marshal(t *testing.T) {
	cap := &captureInvoker{}
	cred := msnrpc.NETLOGON_CREDENTIAL{9, 8, 7, 6, 5, 4, 3, 2}

	if _, _, _, err := functions.NetrServerAuthenticate2(cap, "", "DC$", msnrpc.ServerSecureChannel, "DC", cred, 0x612fffff); err != nil {
		t.Fatalf("NetrServerAuthenticate2: %v", err)
	}
	if cap.opnum != netlogon.OpnumNetrServerAuthenticate2 {
		t.Fatalf("opnum = %d, want %d", cap.opnum, netlogon.OpnumNetrServerAuthenticate2)
	}
	// PrimaryName is NULL -> the stub must begin with a zero referent id.
	if len(cap.stub) < 4 || cap.stub[0] != 0 || cap.stub[1] != 0 || cap.stub[2] != 0 || cap.stub[3] != 0 {
		t.Fatalf("expected NULL PrimaryName referent at start, got %x", cap.stub[:4])
	}
	// NegotiateFlags is the trailing 4-aligned DWORD.
	flags := cap.stub[len(cap.stub)-4:]
	if !bytes.Equal(flags, []byte{0xff, 0xff, 0x2f, 0x61}) {
		t.Fatalf("trailing negotiate flags = %x, want ff ff 2f 61", flags)
	}
	if !bytes.Contains(cap.stub, cred[:]) {
		t.Fatalf("credential %x not found in stub %x", cred[:], cap.stub)
	}
}

// TestNetrServerPasswordSet2Marshal verifies the request carries the 516-octet all-zero
// NL_TRUST_PASSWORD.
func TestNetrServerPasswordSet2Marshal(t *testing.T) {
	cap := &captureInvoker{}
	auth := msnrpc.NETLOGON_AUTHENTICATOR{Credential: msnrpc.NETLOGON_CREDENTIAL{1, 1, 1, 1, 1, 1, 1, 1}, Timestamp: 0}

	if _, _, err := functions.NetrServerPasswordSet2(cap, "DC", "DC$", msnrpc.ServerSecureChannel, "DC", auth, msnrpc.NL_TRUST_PASSWORD{}); err != nil {
		t.Fatalf("NetrServerPasswordSet2: %v", err)
	}
	if cap.opnum != netlogon.OpnumNetrServerPasswordSet2 {
		t.Fatalf("opnum = %d, want %d", cap.opnum, netlogon.OpnumNetrServerPasswordSet2)
	}
	tail := cap.stub[len(cap.stub)-516:]
	for i, b := range tail {
		if b != 0 {
			t.Fatalf("ClearNewPassword octet %d = 0x%02x, want 0", i, b)
		}
	}
}

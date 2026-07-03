package functions_test

import (
	"bytes"
	"encoding/hex"
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

// TestComputeNetlogonCredentialAES pins the credential cryptography ([MS-NRPC] 3.1.4.4.1)
// to a regression vector: AES-128-CFB8 of a fixed challenge under a fixed session key.
func TestComputeNetlogonCredentialAES(t *testing.T) {
	var sk [16]byte
	for i := range sk {
		sk[i] = byte(i)
	}
	challenge := msnrpc.NETLOGON_CREDENTIAL{0, 0, 0, 0, 0x11, 0x11, 0x11, 0}
	got := functions.ComputeNetlogonCredentialAES(challenge, sk)
	if want := "c64020e48fee8f96"; hex.EncodeToString(got[:]) != want {
		t.Fatalf("credential = %x, want %s", got[:], want)
	}
}

// TestComputeSessionKeyAES pins the session-key derivation ([MS-NRPC] 3.1.4.4.1) to a
// regression vector: first 16 octets of HMAC-SHA256(NTOWFv1(password), client || server).
func TestComputeSessionKeyAES(t *testing.T) {
	client := msnrpc.NETLOGON_CREDENTIAL{'1', '2', '3', '4', '5', '6', '7', '8'}
	server := msnrpc.NETLOGON_CREDENTIAL{8, 7, 6, 5, 4, 3, 2, 1}
	got := functions.ComputeSessionKeyAES("Password1", nil, client, server)
	if want := "3a2f0633feed49dcb158b0e72d441508"; hex.EncodeToString(got[:]) != want {
		t.Fatalf("session key = %x, want %s", got[:], want)
	}
}

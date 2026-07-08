package crypto_test

import (
	"encoding/hex"
	"testing"

	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
	"github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc/crypto"
)

var testKey = [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

// TestComputeNetlogonCredentialAES pins the AES credential ([MS-NRPC] 3.1.4.4.1) to a
// regression vector: AES-128-CFB8 of a fixed challenge under a fixed session key.
func TestComputeNetlogonCredentialAES(t *testing.T) {
	challenge := msnrpc.NETLOGON_CREDENTIAL{0, 0, 0, 0, 0x11, 0x11, 0x11, 0}
	got := crypto.ComputeNetlogonCredentialAES(challenge, testKey)
	if want := "c64020e48fee8f96"; hex.EncodeToString(got[:]) != want {
		t.Fatalf("credential = %x, want %s", got[:], want)
	}
}

// TestComputeSessionKeyAES pins the AES session-key derivation ([MS-NRPC] 3.1.4.4.1).
func TestComputeSessionKeyAES(t *testing.T) {
	client := msnrpc.NETLOGON_CREDENTIAL{'1', '2', '3', '4', '5', '6', '7', '8'}
	server := msnrpc.NETLOGON_CREDENTIAL{8, 7, 6, 5, 4, 3, 2, 1}
	got := crypto.ComputeSessionKeyAES("Password1", nil, client, server)
	if want := "3a2f0633feed49dcb158b0e72d441508"; hex.EncodeToString(got[:]) != want {
		t.Fatalf("session key = %x, want %s", got[:], want)
	}
}

// TestComputeSessionKeyStrongKey pins the legacy strong-key session-key derivation ([MS-NRPC]
// 3.1.4.3.1): HMAC-MD5(NTOWFv1, MD5(0000 || client || server)).
func TestComputeSessionKeyStrongKey(t *testing.T) {
	client := msnrpc.NETLOGON_CREDENTIAL{'1', '2', '3', '4', '5', '6', '7', '8'}
	server := msnrpc.NETLOGON_CREDENTIAL{8, 7, 6, 5, 4, 3, 2, 1}
	got := crypto.ComputeSessionKeyStrongKey("Password1", nil, client, server)
	if want := "f69cade46384ab820ec3039702f6ca77"; hex.EncodeToString(got[:]) != want {
		t.Fatalf("strong-key session key = %x, want %s", got[:], want)
	}
}

// TestComputeNetlogonCredential pins the legacy DES credential ([MS-NRPC] 3.1.4.4.2): two
// DES-ECB passes keyed by the two 7-octet halves of the session key.
func TestComputeNetlogonCredential(t *testing.T) {
	input := msnrpc.NETLOGON_CREDENTIAL{1, 2, 3, 4, 5, 6, 7, 8}
	got := crypto.ComputeNetlogonCredential(input, testKey)
	if want := "785fc10827558d95"; hex.EncodeToString(got[:]) != want {
		t.Fatalf("DES credential = %x, want %s", got[:], want)
	}
}

// TestComputeNetlogonAuthenticators pins both authenticator variants ([MS-NRPC] 3.1.4.5):
// stored credential 0x11*8, timestamp 0x11223344, session key 00..0f.
func TestComputeNetlogonAuthenticators(t *testing.T) {
	stored := msnrpc.NETLOGON_CREDENTIAL{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}
	const ts = 0x11223344

	aes := crypto.ComputeNetlogonAuthenticatorAES(stored, ts, testKey)
	if got := hex.EncodeToString(aes.Credential[:]); got != "934de307ed961bf9" {
		t.Errorf("AES authenticator credential = %s, want 934de307ed961bf9", got)
	}
	des := crypto.ComputeNetlogonAuthenticator(stored, ts, testKey)
	if got := hex.EncodeToString(des.Credential[:]); got != "fce444166bf82222" {
		t.Errorf("DES authenticator credential = %s, want fce444166bf82222", got)
	}
	if aes.Timestamp != ts || des.Timestamp != ts {
		t.Errorf("timestamps = %#x/%#x, want %#x", aes.Timestamp, des.Timestamp, ts)
	}
}

// TestAddToCredential pins the low-32-bit little-endian add with overflow ignored and the
// high 4 bytes untouched ([MS-NRPC] 3.1.4.5).
func TestAddToCredential(t *testing.T) {
	in := msnrpc.NETLOGON_CREDENTIAL{0xff, 0xff, 0xff, 0xff, 0xaa, 0xbb, 0xcc, 0xdd}
	got := crypto.AddToCredential(in, 2)
	want := msnrpc.NETLOGON_CREDENTIAL{0x01, 0x00, 0x00, 0x00, 0xaa, 0xbb, 0xcc, 0xdd}
	if got != want {
		t.Errorf("AddToCredential = %x, want %x", got, want)
	}
}

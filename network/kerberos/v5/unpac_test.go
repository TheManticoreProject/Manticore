package kerberos

import (
	"bytes"
	"encoding/asn1"
	"encoding/binary"
	"testing"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/pac"
)

// TestUnPACTheHashGuards covers the pre-conditions UnPACTheHash enforces before
// any network exchange: a TGT is required, and a PKINIT reply key must be
// present (the technique only applies after a certificate logon).
func TestUnPACTheHashGuards(t *testing.T) {
	// No TGT at all.
	c := NewClient("alice", "corp.local", "10.0.0.1").WithPassword("x")
	if _, _, err := c.UnPACTheHash(); err == nil {
		t.Error("expected error without a TGT")
	}

	// A TGT but no PKINIT reply key (a password/hash logon, not certificate).
	c2 := fakeTGTClient(t)
	if _, _, err := c2.UnPACTheHash(); err == nil {
		t.Error("expected error without a PKINIT reply key")
	}
}

// TestExtractWin2KPAC checks the AD-IF-RELEVANT → AD-WIN2K-PAC walk: the PAC
// bytes are recovered from a well-formed wrapper and nil is returned when the
// wrapper or the inner AD-WIN2K-PAC element is absent.
func TestExtractWin2KPAC(t *testing.T) {
	pacBytes := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	inner, err := asn1.Marshal([]messages.AuthorizationData{
		{ADType: adTypeWin2KPAC, ADData: pacBytes},
	})
	if err != nil {
		t.Fatalf("marshal inner AD: %v", err)
	}
	ad := []messages.AuthorizationData{{ADType: adTypeIfRelevant, ADData: inner}}
	if got := extractWin2KPAC(ad); !bytes.Equal(got, pacBytes) {
		t.Errorf("extractWin2KPAC = % X, want % X", got, pacBytes)
	}

	// A non-IF-RELEVANT element is ignored.
	if got := extractWin2KPAC([]messages.AuthorizationData{{ADType: 99, ADData: inner}}); got != nil {
		t.Errorf("expected nil for a non-IF-RELEVANT element, got % X", got)
	}

	// IF-RELEVANT without an AD-WIN2K-PAC element yields nil.
	other, err := asn1.Marshal([]messages.AuthorizationData{{ADType: 42, ADData: []byte{0x01}}})
	if err != nil {
		t.Fatalf("marshal other AD: %v", err)
	}
	if got := extractWin2KPAC([]messages.AuthorizationData{{ADType: adTypeIfRelevant, ADData: other}}); got != nil {
		t.Errorf("expected nil without AD-WIN2K-PAC, got % X", got)
	}
}

// TestPACCredentialInfoRecoversNTHash exercises the credential-recovery half of
// UnPACTheHash (everything after the user-to-user ticket is decrypted): a PAC
// carrying a PAC_CREDENTIAL_INFO encrypted under the PKINIT reply key is parsed,
// decrypted, and the embedded NTLM_SUPPLEMENTAL_CREDENTIAL NT hash is recovered.
// The network user-to-user step itself is not reachable offline.
func TestPACCredentialInfoRecoversNTHash(t *testing.T) {
	replyKey := bytes.Repeat([]byte{0x24}, 32) // AES256 reply key
	replyEType := messages.ETypeAES256CTSHMACSHA196
	wantNT := bytes.Repeat([]byte{0xAB}, 16)
	wantLM := bytes.Repeat([]byte{0xCD}, 16)

	// PAC_CREDENTIAL_DATA: NDR type-serialization framing, then the trailing
	// NTLM_SUPPLEMENTAL_CREDENTIAL as a conformant array (MaximumCount=40).
	ntlm := make([]byte, 40)
	binary.LittleEndian.PutUint32(ntlm[0:], 0) // Version
	binary.LittleEndian.PutUint32(ntlm[4:], 0x1|0x2)
	copy(ntlm[8:], wantLM)
	copy(ntlm[24:], wantNT)
	var credData []byte
	credData = append(credData, bytes.Repeat([]byte{0x00}, 48)...)
	count := make([]byte, 4)
	binary.LittleEndian.PutUint32(count, 40)
	credData = append(credData, count...)
	credData = append(credData, ntlm...)

	// Encrypt PAC_CREDENTIAL_DATA with the reply key (key usage KERB_NON_KERB_SALT).
	enc, err := kerbcrypto.Encrypt(replyEType, replyKey, kerbcrypto.KeyUsageKerbNonKerbSalt, credData)
	if err != nil {
		t.Fatalf("encrypt PAC_CREDENTIAL_DATA: %v", err)
	}

	// PAC_CREDENTIAL_INFO: Version(0), EncryptionType, SerializedData(ciphertext).
	ci := make([]byte, 8)
	binary.LittleEndian.PutUint32(ci[0:], 0)
	binary.LittleEndian.PutUint32(ci[4:], uint32(replyEType))
	ci = append(ci, enc...)

	p := &pac.PAC{Version: 0, Buffers: []pac.Buffer{{Type: pac.BufferCredentials, Data: ci}}}
	pacBytes, err := p.Marshal()
	if err != nil {
		t.Fatalf("PAC.Marshal: %v", err)
	}

	// Drive the same parse/decrypt chain UnPACTheHash runs on the U2U ticket's PAC.
	parsed, err := pac.Parse(pacBytes)
	if err != nil {
		t.Fatalf("pac.Parse: %v", err)
	}
	buf, ok := parsed.Buffer(pac.BufferCredentials)
	if !ok {
		t.Fatal("PAC has no PAC_CREDENTIAL_INFO buffer")
	}
	info, err := pac.ParseCredentialInfo(buf.Data)
	if err != nil {
		t.Fatalf("ParseCredentialInfo: %v", err)
	}
	plain, err := info.DecryptCredentialData(replyKey)
	if err != nil {
		t.Fatalf("DecryptCredentialData: %v", err)
	}
	ntlmCred, err := pac.ParseNTLMSupplementalCredential(plain)
	if err != nil {
		t.Fatalf("ParseNTLMSupplementalCredential: %v", err)
	}
	if !bytes.Equal(ntlmCred.NTHash, wantNT) {
		t.Errorf("recovered NT hash = %x, want %x", ntlmCred.NTHash, wantNT)
	}
	if !bytes.Equal(ntlmCred.LMHash, wantLM) {
		t.Errorf("recovered LM hash = %x, want %x", ntlmCred.LMHash, wantLM)
	}

	// A wrong reply key must fail decryption (integrity check).
	if _, err := info.DecryptCredentialData(bytes.Repeat([]byte{0x99}, 32)); err == nil {
		t.Error("expected decryption failure with a wrong reply key")
	}
}

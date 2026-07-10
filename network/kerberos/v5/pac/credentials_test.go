package pac

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestParseCredentialInfo(t *testing.T) {
	buf := make([]byte, 8+16)
	binary.LittleEndian.PutUint32(buf[0:], 0)  // Version
	binary.LittleEndian.PutUint32(buf[4:], 18) // EncryptionType = AES256
	copy(buf[8:], bytes.Repeat([]byte{0xAA}, 16))
	ci, err := ParseCredentialInfo(buf)
	if err != nil {
		t.Fatalf("ParseCredentialInfo: %v", err)
	}
	if ci.Version != 0 || ci.EncryptionType != 18 {
		t.Fatalf("header mismatch: v=%d et=%d", ci.Version, ci.EncryptionType)
	}
	if len(ci.SerializedData) != 16 {
		t.Fatalf("SerializedData len %d, want 16", len(ci.SerializedData))
	}
}

// TestParseNTLMSupplementalCredential builds a decrypted PAC_CREDENTIAL_DATA
// carrying a single "NTLM" credential and verifies the NT hash is recovered.
func TestParseNTLMSupplementalCredential(t *testing.T) {
	nt := bytes.Repeat([]byte{0x11}, 16)
	lm := bytes.Repeat([]byte{0x22}, 16)

	// NTLM_SUPPLEMENTAL_CREDENTIAL: Version(0), Flags(LM|NT), LM[16], NT[16].
	ntlm := make([]byte, 40)
	binary.LittleEndian.PutUint32(ntlm[0:], 0)
	binary.LittleEndian.PutUint32(ntlm[4:], ntlmCredentialLMPresent|ntlmCredentialNTPresent)
	copy(ntlm[8:], lm)
	copy(ntlm[24:], nt)

	// Prepend some plausible NDR framing plus the conformant array MaximumCount
	// (=40) that precedes the trailing NTLM cred.
	var blob []byte
	blob = append(blob, bytes.Repeat([]byte{0x00}, 48)...) // header + inline fields
	count := make([]byte, 4)
	binary.LittleEndian.PutUint32(count, 40)
	blob = append(blob, count...)
	blob = append(blob, ntlm...)

	cred, err := ParseNTLMSupplementalCredential(blob)
	if err != nil {
		t.Fatalf("ParseNTLMSupplementalCredential: %v", err)
	}
	if !bytes.Equal(cred.NTHash, nt) {
		t.Fatalf("NT hash mismatch: got %x", cred.NTHash)
	}
	if !bytes.Equal(cred.LMHash, lm) {
		t.Fatalf("LM hash mismatch: got %x", cred.LMHash)
	}
}

func TestParseNTLMSupplementalCredentialNTOnly(t *testing.T) {
	nt := bytes.Repeat([]byte{0x33}, 16)
	ntlm := make([]byte, 40)
	binary.LittleEndian.PutUint32(ntlm[4:], ntlmCredentialNTPresent)
	copy(ntlm[24:], nt)
	count := make([]byte, 4)
	binary.LittleEndian.PutUint32(count, 40)
	blob := append(append(bytes.Repeat([]byte{0}, 32), count...), ntlm...)

	cred, err := ParseNTLMSupplementalCredential(blob)
	if err != nil {
		t.Fatalf("ParseNTLMSupplementalCredential: %v", err)
	}
	if cred.LMHash != nil {
		t.Fatal("LM hash should be nil when not flagged present")
	}
	if !bytes.Equal(cred.NTHash, nt) {
		t.Fatalf("NT hash mismatch: got %x", cred.NTHash)
	}
}

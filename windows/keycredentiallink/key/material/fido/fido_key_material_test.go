package fido

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestFIDOKeyMaterialRoundTrip(t *testing.T) {
	authBytes := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	certDER := []byte{0x30, 0x82, 0x01, 0x00}

	f := &FIDOKeyMaterial{
		Version:     1,
		AuthData:    base64.StdEncoding.EncodeToString(authBytes),
		X5C:         []string{base64.StdEncoding.EncodeToString(certDER)},
		DisplayName: "YubiKey 5 NFC",
	}

	data, err := f.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Verify it's valid JSON.
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Marshal produced invalid JSON: %v", err)
	}

	f2 := &FIDOKeyMaterial{}
	n, err := f2.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if n != len(data) {
		t.Errorf("bytesRead = %d, want %d", n, len(data))
	}
	if f2.Version != 1 {
		t.Errorf("Version = %d, want 1", f2.Version)
	}
	if f2.DisplayName != "YubiKey 5 NFC" {
		t.Errorf("DisplayName = %q, want %q", f2.DisplayName, "YubiKey 5 NFC")
	}
	if f2.AuthData != f.AuthData {
		t.Errorf("AuthData mismatch")
	}
	if len(f2.X5C) != 1 {
		t.Fatalf("X5C length = %d, want 1", len(f2.X5C))
	}
}

func TestAuthenticatorData(t *testing.T) {
	raw := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	f := &FIDOKeyMaterial{
		AuthData: base64.StdEncoding.EncodeToString(raw),
	}
	got, err := f.AuthenticatorData()
	if err != nil {
		t.Fatalf("AuthenticatorData: %v", err)
	}
	if len(got) != len(raw) {
		t.Fatalf("len = %d, want %d", len(got), len(raw))
	}
	for i := range raw {
		if got[i] != raw[i] {
			t.Errorf("byte %d = 0x%02x, want 0x%02x", i, got[i], raw[i])
		}
	}
}

func TestAttestationCertificates(t *testing.T) {
	cert1 := []byte{0x30, 0x01}
	cert2 := []byte{0x30, 0x02}
	f := &FIDOKeyMaterial{
		X5C: []string{
			base64.StdEncoding.EncodeToString(cert1),
			base64.StdEncoding.EncodeToString(cert2),
		},
	}
	certs, err := f.AttestationCertificates()
	if err != nil {
		t.Fatalf("AttestationCertificates: %v", err)
	}
	if len(certs) != 2 {
		t.Fatalf("len = %d, want 2", len(certs))
	}
	if certs[0][1] != 0x01 || certs[1][1] != 0x02 {
		t.Errorf("cert data mismatch")
	}
}

func TestAttestationCertificatesBadBase64(t *testing.T) {
	f := &FIDOKeyMaterial{X5C: []string{"not-valid-base64!!!"}}
	if _, err := f.AttestationCertificates(); err == nil {
		t.Error("expected error for bad base64")
	}
}

func TestUnmarshalBadJSON(t *testing.T) {
	f := &FIDOKeyMaterial{}
	if _, err := f.Unmarshal([]byte("not json")); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFingerprint(t *testing.T) {
	f := &FIDOKeyMaterial{Version: 1, DisplayName: "TestKey"}
	if got := f.Fingerprint(); got != "FIDO_KEY_MATERIAL:v1:TestKey" {
		t.Errorf("Fingerprint = %q", got)
	}
}

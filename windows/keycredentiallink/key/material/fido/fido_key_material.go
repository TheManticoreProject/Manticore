// Package fido implements the KEY_USAGE_FIDO key material variant stored in a
// KEYCREDENTIALLINK_ENTRY. Unlike the RSA/CNG blob used by NGC keys, FIDO key
// material is a UTF-8 JSON object containing WebAuthn Authenticator Data and an
// attestation certificate chain.
//
// Reference: [MS-ADTS] 2.2.20.1
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/d5948ab9-8993-4066-a175-851af361ea7f
package fido

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt"
)

var _ bcrypt.KeyMaterial = (*FIDOKeyMaterial)(nil)

// FIDOKeyMaterial is the KEY_USAGE_FIDO key material stored in a
// KEYCREDENTIALLINK_ENTRY's KeyMaterial value. The entry contains the raw UTF-8
// JSON bytes of this structure.
type FIDOKeyMaterial struct {
	// Version is the structure version number.
	Version int `json:"version"`

	// AuthData is base64-encoded WebAuthn Authenticator Data ([W3C-WebAuthPKC1]
	// 6.1). It embeds the AAGUID, credential ID, and COSE-encoded public key.
	AuthData string `json:"authData"`

	// X5C is the base64-encoded X.509 attestation certificate chain.
	X5C []string `json:"x5c"`

	// DisplayName is a human-readable label for the credential.
	DisplayName string `json:"displayName"`
}

// Unmarshal parses FIDO key material from the raw bytes of a KeyMaterial entry.
func (f *FIDOKeyMaterial) Unmarshal(data []byte) (int, error) {
	if err := json.Unmarshal(data, f); err != nil {
		return 0, fmt.Errorf("fido key material: %w", err)
	}
	return len(data), nil
}

// Marshal returns the UTF-8 JSON encoding of the FIDO key material.
func (f *FIDOKeyMaterial) Marshal() ([]byte, error) {
	return json.Marshal(f)
}

// AuthenticatorData decodes the base64 AuthData field into raw bytes.
func (f *FIDOKeyMaterial) AuthenticatorData() ([]byte, error) {
	return base64.StdEncoding.DecodeString(f.AuthData)
}

// AttestationCertificates decodes the base64 X5C entries into raw DER
// certificate bytes.
func (f *FIDOKeyMaterial) AttestationCertificates() ([][]byte, error) {
	certs := make([][]byte, len(f.X5C))
	for i, s := range f.X5C {
		der, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("x5c[%d]: %w", i, err)
		}
		certs[i] = der
	}
	return certs, nil
}

// Describe prints a detailed description of the FIDO key material.
func (f *FIDOKeyMaterial) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mFIDO_KEY_MATERIAL\x1b[0m>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mVersion\x1b[0m: %d\n", indentPrompt, f.Version)
	fmt.Printf("%s │ \x1b[93mDisplayName\x1b[0m: %s\n", indentPrompt, f.DisplayName)
	if len(f.AuthData) > 0 {
		authLen := len(f.AuthData)
		preview := f.AuthData
		if authLen > 40 {
			preview = f.AuthData[:40] + "..."
		}
		fmt.Printf("%s │ \x1b[93mAuthData\x1b[0m: %s (%d chars base64)\n", indentPrompt, preview, authLen)
	}
	fmt.Printf("%s │ \x1b[93mX5C\x1b[0m: %d certificate(s)\n", indentPrompt, len(f.X5C))
	fmt.Printf("%s └───\n", indentPrompt)
}

// Fingerprint returns a string fingerprint of the FIDO key material.
func (f *FIDOKeyMaterial) Fingerprint() string {
	return fmt.Sprintf("FIDO_KEY_MATERIAL:v%d:%s", f.Version, f.DisplayName)
}

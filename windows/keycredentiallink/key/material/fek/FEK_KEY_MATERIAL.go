package fek

import (
	"errors"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/material/fek/blob"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/material/fek/headers"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/material/fek/magic"
)

// FEK_KEY_MATERIAL represents the KEY_USAGE_FEK key material stored in a
// KEYCREDENTIALLINK_ENTRY. The buffer is a combination of an RSA 2048 public
// key (RFC8017) and an AES-256 KDF key, prefixed by a single version byte whose
// value MUST be 1 (see FekKeyVersion in CUSTOM_KEY_INFORMATION).
//
// See:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/735fd27a-3f22-4926-93f9-0298bb67a84b
type FEK_KEY_MATERIAL struct {
	// Version is the 1-byte version tag. MUST be FEK_KEY_VERSION_1 (0x01).
	Version magic.FEK_KEY_MATERIAL_VERSION

	// Header describes the sizes of the RSA and AES-256 KDF key material.
	Header headers.FEK_KEY_MATERIAL_HEADER

	// Content contains the RSA public exponent, modulus, and AES-256 KDF key.
	Content blob.FEK_KEY_MATERIAL_BLOB
}

// Unmarshal parses the provided byte slice into the FEK_KEY_MATERIAL structure.
//
// Parameters:
// - value: A byte slice containing the raw FEK key material to be parsed.
//
// Returns:
// - The number of bytes read from the byte slice.
// - An error if the parsing fails, otherwise nil.
func (k *FEK_KEY_MATERIAL) Unmarshal(value []byte) (int, error) {
	if len(value) < 17 {
		return 0, errors.New("buffer too small for FEK_KEY_MATERIAL, header too short (at least 17 bytes are required for version and header)")
	}

	bytesRead := 0

	// Unmarshalling version
	bytesReadVersion, err := k.Version.Unmarshal(value[bytesRead:])
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal FEK key material version: %w", err)
	}
	if k.Version.Version != magic.FEK_KEY_VERSION_1 {
		return 0, fmt.Errorf("failed to unmarshal FEK key material version: invalid version: 0x%02x", k.Version.Version)
	}
	bytesRead += bytesReadVersion

	// Unmarshalling header
	bytesReadHeader, err := k.Header.Unmarshal(value[bytesRead:])
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal FEK key material header: %w", err)
	}
	bytesRead += bytesReadHeader

	// Unmarshalling content
	bytesReadContent, err := k.Content.Unmarshal(k.Header, value[bytesRead:])
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal FEK key material content: %w", err)
	}
	bytesRead += bytesReadContent

	return bytesRead, nil
}

// Marshal returns the raw bytes of the FEK_KEY_MATERIAL structure.
//
// Returns:
// - A byte slice representing the raw bytes of the FEK_KEY_MATERIAL structure.
// - An error if the conversion fails.
func (k *FEK_KEY_MATERIAL) Marshal() ([]byte, error) {
	marshalledData := []byte{}

	// Marshalling version
	k.Version.Version = magic.FEK_KEY_VERSION_1
	marshalledVersion, err := k.Version.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, marshalledVersion...)

	// Marshalling header
	marshalledHeader, err := k.Header.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, marshalledHeader...)

	// Marshalling content
	marshalledContent, err := k.Content.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, marshalledContent...)

	return marshalledData, nil
}

// Describe prints a detailed description of the FEK_KEY_MATERIAL structure.
//
// Parameters:
// - indent: An integer representing the indentation level for the printed output.
func (k *FEK_KEY_MATERIAL) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mFEK_KEY_MATERIAL\x1b[0m>\n", indentPrompt)
	k.Version.Describe(indent + 1)
	k.Header.Describe(indent + 1)
	k.Content.Describe(indent + 1)
	fmt.Printf("%s └───\n", indentPrompt)
}

// Equal checks if two FEK_KEY_MATERIAL structures are equal.
//
// Parameters:
// - other: The FEK_KEY_MATERIAL structure to compare to.
//
// Returns:
// - True if the two FEK_KEY_MATERIAL structures are equal, false otherwise.
func (k *FEK_KEY_MATERIAL) Equal(other *FEK_KEY_MATERIAL) bool {
	return k.Version.Equal(&other.Version) && k.Header.Equal(&other.Header) && k.Content.Equal(&other.Content)
}

// Fingerprint returns the fingerprint of the FEK_KEY_MATERIAL structure.
//
// Returns:
// - A string representing the fingerprint of the FEK_KEY_MATERIAL structure.
func (k *FEK_KEY_MATERIAL) Fingerprint() string {
	return fmt.Sprintf("FEK_KEY_MATERIAL:0x%x:0x%x:0x%x", k.Content.PublicExponent, k.Content.Modulus, k.Content.AESKDFKey)
}

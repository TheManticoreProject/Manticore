package keys

import (
	"crypto/rsa"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/blob"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/headers"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/magic"
)

type BCRYPT_RSA_PUBLIC_KEY struct {
	// Magic is the magic signature of the key.
	Magic magic.BCRYPT_KEY_BLOB

	// Header is the header of the key.
	Header headers.BCRYPT_RSA_KEY_BLOB

	// Content is the content of the key.
	Content blob.BCRYPT_RSA_PUBLIC_BLOB
}

// NewBCRYPT_RSA_PUBLIC_KEY builds a BCRYPT_RSA_PUBLIC_KEY ("RSA1" public blob)
// from a crypto/rsa public key. It is the inverse of ExportPEM/ExportDER and
// produces the CNG key-material bytes (via Marshal) embedded, for example, in a
// KEYCREDENTIALLINK_BLOB. The modulus and public exponent are stored big-endian
// with their minimal (leading-zero-free) length, matching the CbModulus /
// CbPublicExp header fields.
func NewBCRYPT_RSA_PUBLIC_KEY(pub *rsa.PublicKey) *BCRYPT_RSA_PUBLIC_KEY {
	modulus := pub.N.Bytes()
	exponent := big.NewInt(int64(pub.E)).Bytes()

	return &BCRYPT_RSA_PUBLIC_KEY{
		Magic: magic.BCRYPT_KEY_BLOB{Magic: magic.BCRYPT_RSAPUBLIC_MAGIC},
		Header: headers.BCRYPT_RSA_KEY_BLOB{
			BitLength:   uint32(pub.N.BitLen()),
			CbPublicExp: uint32(len(exponent)),
			CbModulus:   uint32(len(modulus)),
		},
		Content: blob.BCRYPT_RSA_PUBLIC_BLOB{
			PublicExponent: exponent,
			Modulus:        modulus,
		},
	}
}

// Unmarshal parses the provided byte slice into the BCRYPT_RSA_PUBLIC_KEY structure.
//
// Parameters:
// - value: A byte slice containing the raw RSA public key to be parsed.
//
// Returns:
// - The number of bytes read from the byte slice.
// - An error if the parsing fails, otherwise nil.
//
// Note:
// The function expects the byte slice to follow the RSA public key format, starting with the BCRYPT_RSA_KEY_BLOB header.
// It extracts the public exponent and modulus from the byte slice and stores them in the BCRYPT_RSA_PUBLIC_KEY structure.
func (k *BCRYPT_RSA_PUBLIC_KEY) Unmarshal(value []byte) (int, error) {
	if len(value) < 24 {
		return 0, errors.New("buffer too small for BCRYPT_RSA_PUBLIC_KEY, header too short (at least 24 bytes are required for magic and header)")
	}

	bytesRead := 0

	// Unmarshalling magic
	bytesReadMagic, err := k.Magic.Unmarshal(value[bytesRead:])
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal RSA public key magic: %w", err)
	}
	if k.Magic.Magic != magic.BCRYPT_RSAPUBLIC_MAGIC {
		return 0, fmt.Errorf("failed to unmarshal RSA public key magic: invalid magic: 0x%08x", k.Magic.Magic)
	}
	bytesRead += bytesReadMagic

	// Unmarshalling header
	bytesReadHeader, err := k.Header.Unmarshal(value[bytesRead:])
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal RSA public key header: %w", err)
	}
	bytesRead += bytesReadHeader

	// Unmarshalling content
	bytesReadContent, err := k.Content.Unmarshal(k.Header, value[bytesRead:])
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal RSA public key content: %w", err)
	}
	bytesRead += bytesReadContent

	return bytesRead, nil
}

// Marshal returns the raw bytes of the BCRYPT_RSA_PUBLIC_KEY structure.
//
// Returns:
// - A byte slice representing the raw bytes of the BCRYPT_RSA_PUBLIC_KEY structure.
func (k *BCRYPT_RSA_PUBLIC_KEY) Marshal() ([]byte, error) {
	marshalledData := []byte{}

	// Marshalling magic
	k.Magic.Magic = magic.BCRYPT_RSAPUBLIC_MAGIC
	marshalledMagic, err := k.Magic.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, marshalledMagic...)

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

// Describe prints a detailed description of the BCRYPT_RSA_PUBLIC_KEY structure.
//
// Parameters:
// - indent: An integer representing the indentation level for the printed output.
//
// Note:
// The function prints the Header and Data of the BCRYPT_RSA_PUBLIC_KEY structure.
// The output is formatted with the specified indentation level to improve readability.
func (k *BCRYPT_RSA_PUBLIC_KEY) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mBCRYPT_RSA_PUBLIC_KEY\x1b[0m>\n", indentPrompt)
	k.Magic.Describe(indent + 1)
	k.Header.Describe(indent + 1)
	k.Content.Describe(indent + 1)
	fmt.Printf("%s └───\n", indentPrompt)
}

// Equal checks if two BCRYPT_RSA_PUBLIC_KEY structures are equal.
//
// Parameters:
// - other: The BCRYPT_RSA_PUBLIC_KEY structure to compare to.
//
// Returns:
// - True if the two BCRYPT_RSA_PUBLIC_KEY structures are equal, false otherwise.
func (k *BCRYPT_RSA_PUBLIC_KEY) Equal(other *BCRYPT_RSA_PUBLIC_KEY) bool {
	return k.Magic.Equal(&other.Magic) && k.Header.Equal(&other.Header) && k.Content.Equal(&other.Content)
}

// Fingerprint returns the fingerprint of the BCRYPT_RSA_PUBLIC_KEY structure.
//
// Parameters:
// - key: The BCRYPT_RSA_PUBLIC_KEY structure to get the fingerprint of.
//
// Returns:
// - A string representing the fingerprint of the BCRYPT_RSA_PUBLIC_KEY structure.
func (key *BCRYPT_RSA_PUBLIC_KEY) Fingerprint() string {
	return fmt.Sprintf("BCRYPT_RSA_PUBLIC_KEY:0x%x:0x%x", key.Content.PublicExponent, key.Content.Modulus)
}

// ExportPEM exports the RSA public key in PEM format.
//
// Returns:
// - A byte slice containing the PEM-encoded RSA public key.
// - An error if encoding fails.
func (key *BCRYPT_RSA_PUBLIC_KEY) ExportPEM() ([]byte, error) {
	// Build rsa.PublicKey manually from modulus and public exponent

	publicExponent := new(big.Int).SetBytes(key.Content.PublicExponent)
	// ASN.1 encode the public key in PKCS#1 format
	type rsaPublicKey struct {
		N *big.Int
		E *big.Int
	}
	n := new(big.Int).SetBytes(key.Content.Modulus)
	pk := rsaPublicKey{N: n, E: publicExponent}
	asn1Bytes, err := asn1.Marshal(pk)
	if err != nil {
		return nil, err
	}

	// Build the ASN.1 SubjectPublicKeyInfo
	// See RFC 5280, section 4.1, 4.1.2.7, and RFC 3447.
	var spkiAlgoID = []byte{
		0x30, 0x0d, // SEQUENCE, 13 bytes
		0x06, 0x09, // OID (1.2.840.113549.1.1.1, rsaEncryption)
		0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x01, 0x01,
		0x05, 0x00, // NULL
	}
	// SubjectPublicKeyInfo ::= SEQUENCE {
	//    algorithm          AlgorithmIdentifier,
	//    subjectPublicKey   BIT STRING
	// }
	
	// Marshal the BIT STRING
	bitString := asn1.BitString{
		Bytes:     asn1Bytes,
		BitLength: len(asn1Bytes) * 8,
	}
	bitStringBytes, err := asn1.Marshal(bitString)
	if err != nil {
		return nil, err
	}
	
	spkSeq := asn1.RawValue{
		Class:      asn1.ClassUniversal,
		Tag:        asn1.TagSequence,
		IsCompound: true,
		Bytes:      append(spkiAlgoID, bitStringBytes...),
	}
	der, err := asn1.Marshal(spkSeq)
	if err != nil {
		return nil, err
	}

	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	}
	return pem.EncodeToMemory(block), nil
}

// ExportDER exports the RSA public key in DER format (SubjectPublicKeyInfo/PKIX).
//
// Returns:
// - A byte slice containing the DER-encoded RSA public key.
// - An error if encoding fails.
func (key *BCRYPT_RSA_PUBLIC_KEY) ExportDER() ([]byte, error) {
	// Same logic as ExportPEM, but return the DER bytes for SubjectPublicKeyInfo.

	publicExponent := new(big.Int).SetBytes(key.Content.PublicExponent)
	n := new(big.Int).SetBytes(key.Content.Modulus)
	type rsaPublicKey struct {
		N *big.Int
		E *big.Int
	}
	pk := rsaPublicKey{N: n, E: publicExponent}
	asn1Bytes, err := asn1.Marshal(pk)
	if err != nil {
		return nil, err
	}

	var spkiAlgoID = []byte{
		0x30, 0x0d, // SEQUENCE, 13 bytes
		0x06, 0x09,
		0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x01, 0x01,
		0x05, 0x00,
	}
	
	// Marshal the BIT STRING
	bitString := asn1.BitString{
		Bytes:     asn1Bytes,
		BitLength: len(asn1Bytes) * 8,
	}
	bitStringBytes, err := asn1.Marshal(bitString)
	if err != nil {
		return nil, err
	}
	
	spkSeq := asn1.RawValue{
		Class:      asn1.ClassUniversal,
		Tag:        asn1.TagSequence,
		IsCompound: true,
		Bytes:      append(spkiAlgoID, bitStringBytes...),
	}
	der, err := asn1.Marshal(spkSeq)
	if err != nil {
		return nil, err
	}
	return der, nil
}

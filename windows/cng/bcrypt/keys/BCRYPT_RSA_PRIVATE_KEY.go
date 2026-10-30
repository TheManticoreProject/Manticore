package keys

import (
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

type BCRYPT_RSA_PRIVATE_KEY struct {
	// Magic is the magic signature of the key.
	Magic magic.BCRYPT_KEY_BLOB

	// Header is the header of the key.
	Header headers.BCRYPT_RSA_KEY_BLOB

	// Content is the content of the key.
	Content blob.BCRYPT_RSA_PRIVATE_BLOB
}

// Unmarshal parses the provided byte slice into the BCRYPT_RSA_PRIVATE_KEY structure.
//
// Parameters:
// - value: A byte slice containing the raw RSA private key to be parsed.
//
// Returns:
// - The number of bytes read from the byte slice.
// - An error if the parsing fails, otherwise nil.
//
// Note:
// The function expects the byte slice to follow the RSA private key format, starting with the BCRYPT_RSA_KEY_BLOB header.
// It extracts the public exponent, modulus, prime1, and prime2 from the byte slice and stores them in the BCRYPT_RSA_PRIVATE_KEY structure.
func (k *BCRYPT_RSA_PRIVATE_KEY) Unmarshal(value []byte) (int, error) {
	if len(value) < 24 {
		return 0, errors.New("buffer too small for BCRYPT_RSA_PRIVATE_KEY, header too short (at least 24 bytes are required)")
	}

	bytesRead := 0

	// Unmarshalling magic
	bytesReadMagic, err := k.Magic.Unmarshal(value[:4])
	if err != nil {
		return 0, err
	}
	if k.Magic.Magic != magic.BCRYPT_RSAPRIVATE_MAGIC {
		return 0, fmt.Errorf("invalid RSA private key magic: 0x%08x", k.Magic.Magic)
	}
	bytesRead += bytesReadMagic

	// Unmarshalling header
	bytesReadHeader, err := k.Header.Unmarshal(value[bytesRead:])
	if err != nil {
		return 0, err
	}
	bytesRead += bytesReadHeader

	// Unmarshalling content
	bytesReadContent, err := k.Content.Unmarshal(k.Header, value[bytesRead:])
	if err != nil {
		return 0, err
	}
	bytesRead += bytesReadContent

	return bytesRead, nil
}

// Marshal returns the raw bytes of the BCRYPT_RSA_PRIVATE_KEY structure.
//
// Returns:
// - A byte slice representing the raw bytes of the BCRYPT_RSA_PRIVATE_KEY structure.
func (k *BCRYPT_RSA_PRIVATE_KEY) Marshal() ([]byte, error) {
	marshalledData := []byte{}

	// Marshalling magic
	k.Magic.Magic = magic.BCRYPT_RSAPRIVATE_MAGIC
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

// Describe prints a detailed description of the BCRYPT_RSA_PRIVATE_KEY structure.
//
// Parameters:
// - indent: An integer representing the indentation level for the printed output.
//
// Note:
// The function prints the Header and Data of the BCRYPT_RSA_PRIVATE_KEY structure.
// The output is formatted with the specified indentation level to improve readability.
func (k *BCRYPT_RSA_PRIVATE_KEY) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mBCRYPT_RSA_PRIVATE_KEY\x1b[0m>\n", indentPrompt)
	k.Magic.Describe(indent + 1)
	k.Header.Describe(indent + 1)
	k.Content.Describe(indent + 1)
	fmt.Printf("%s└───\n", indentPrompt)
}

// Fingerprint returns the fingerprint of the BCRYPT_RSA_PUBLIC_KEY structure.
//
// Parameters:
// - key: The BCRYPT_RSA_PRIVATE_KEY structure to get the fingerprint of.
//
// Returns:
// - A string representing the fingerprint of the BCRYPT_RSA_PRIVATE_KEY structure.
func (key *BCRYPT_RSA_PRIVATE_KEY) Fingerprint() string {
	return fmt.Sprintf("BCRYPT_RSA_PRIVATE_KEY:0x%x:0x%x:0x%x:0x%x", key.Content.PublicExponent, key.Content.Modulus, key.Content.Prime1, key.Content.Prime2)
}

// ExportPEM exports the RSA private key in PEM format (PKCS#1).
//
// Returns:
// - A byte slice containing the PEM-encoded RSA private key.
// - An error if encoding fails.
func (key *BCRYPT_RSA_PRIVATE_KEY) ExportPEM() ([]byte, error) {
	der, err := key.ExportDER()
	if err != nil {
		return nil, err
	}
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: der,
	}
	return pem.EncodeToMemory(block), nil
}

// ExportDER exports the RSA private key in DER format (PKCS#1 RSAPrivateKey).
//
// The structure encoded is:
//
//	RSAPrivateKey ::= SEQUENCE {
//	  version           Version,
//	  modulus           INTEGER,  -- n
//	  publicExponent    INTEGER,  -- e
//	  privateExponent   INTEGER,  -- d
//	  prime1            INTEGER,  -- p
//	  prime2            INTEGER,  -- q
//	  exponent1         INTEGER,  -- d mod (p-1)
//	  exponent2         INTEGER,  -- d mod (q-1)
//	  coefficient       INTEGER   -- (inverse of q) mod p
//	}
//
// Returns:
// - A byte slice containing the DER-encoded RSA private key.
// - An error if encoding fails.
func (key *BCRYPT_RSA_PRIVATE_KEY) ExportDER() ([]byte, error) {
	// Convert components to big.Int
	n := new(big.Int).SetBytes(key.Content.Modulus)
	e := new(big.Int).SetBytes(key.Content.PublicExponent)
	p := new(big.Int).SetBytes(key.Content.Prime1)
	q := new(big.Int).SetBytes(key.Content.Prime2)
	if n.Sign() == 0 || e.Sign() == 0 || p.Sign() == 0 || q.Sign() == 0 {
		return nil, errors.New("invalid RSA private key components")
	}

	// Primes of 1 would make p-1 or q-1 zero, and the CRT reductions below
	// would divide by zero
	one := big.NewInt(1)
	if p.Cmp(one) <= 0 || q.Cmp(one) <= 0 {
		return nil, errors.New("invalid RSA private key components: primes must be greater than 1")
	}
	pMinus1 := new(big.Int).Sub(p, one)
	qMinus1 := new(big.Int).Sub(q, one)
	phi := new(big.Int).Mul(pMinus1, qMinus1)

	// Compute private exponent d = e^{-1} mod phi(n)
	d := new(big.Int).ModInverse(e, phi)
	if d == nil {
		return nil, errors.New("failed to compute private exponent: e has no inverse modulo phi(n)")
	}

	// CRT parameters
	dp := new(big.Int).Mod(d, pMinus1)
	dq := new(big.Int).Mod(d, qMinus1)
	qi := new(big.Int).ModInverse(q, p) // coefficient = q^{-1} mod p
	if qi == nil {
		return nil, errors.New("failed to compute CRT coefficient")
	}

	// ASN.1 PKCS#1 RSAPrivateKey
	type rsaPrivateKey struct {
		Version int
		N       *big.Int
		E       *big.Int
		D       *big.Int
		P       *big.Int
		Q       *big.Int
		Dp      *big.Int
		Dq      *big.Int
		Qi      *big.Int
	}
	priv := rsaPrivateKey{
		Version: 0,
		N:       n,
		E:       e,
		D:       d,
		P:       p,
		Q:       q,
		Dp:      dp,
		Dq:      dq,
		Qi:      qi,
	}
	return asn1.Marshal(priv)
}

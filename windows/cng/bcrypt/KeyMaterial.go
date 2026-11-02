package bcrypt

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/magic"
)

// KeyMaterial is an interface that represents a key material in memory.
//
// It provides methods for unmarshalling, marshalling, and comparing key materials.
type KeyMaterial interface {
	Unmarshal(data []byte) (int, error)
	Marshal() ([]byte, error)
	Describe(indent int)
	Fingerprint() string
}

// UnmarshalKeyMaterial unmarshals the key material from the provided byte slice.
//
// Parameters:
// - data: A byte slice containing the raw key material to be parsed.
//
// Returns:
// - A KeyMaterial object representing the key material.
// - An error if the unmarshalling fails, otherwise nil.
//
// Note:
// The blob type is selected by the leading magic: RSA public and private
// (BCRYPT_RSAKEY_BLOB), ECDH and ECDSA P-256/P-384/P-521 public and private
// (BCRYPT_ECCKEY_BLOB), and DSA public and private (BCRYPT_DSA_KEY_BLOB, the
// FIPS 186-2 form). Windows Hello for Business can register ECC keys in
// msDS-KeyCredentialLink.
func UnmarshalKeyMaterial(data []byte) (KeyMaterial, int, error) {
	blobMagic := magic.BCRYPT_KEY_BLOB{}
	_, err := blobMagic.Unmarshal(data)
	if err != nil {
		return nil, 0, err
	}

	switch blobMagic.Magic {

	case magic.BCRYPT_RSAPUBLIC_MAGIC:
		blob := keys.BCRYPT_RSA_PUBLIC_KEY{}
		bytesRead, err := blob.Unmarshal(data)
		if err != nil {
			return nil, 0, err
		}
		return &blob, bytesRead, nil

	case magic.BCRYPT_RSAPRIVATE_MAGIC:
		blob := keys.BCRYPT_RSA_PRIVATE_KEY{}
		bytesRead, err := blob.Unmarshal(data)
		if err != nil {
			return nil, 0, err
		}
		return &blob, bytesRead, nil

	case magic.BCRYPT_ECDH_PUBLIC_P256_MAGIC, magic.BCRYPT_ECDH_PUBLIC_P384_MAGIC, magic.BCRYPT_ECDH_PUBLIC_P521_MAGIC,
		magic.BCRYPT_ECDSA_PUBLIC_P256_MAGIC, magic.BCRYPT_ECDSA_PUBLIC_P384_MAGIC, magic.BCRYPT_ECDSA_PUBLIC_P521_MAGIC:
		blob := keys.BCRYPT_ECC_PUBLIC_KEY{}
		bytesRead, err := blob.Unmarshal(data)
		if err != nil {
			return nil, 0, err
		}
		return &blob, bytesRead, nil

	case magic.BCRYPT_ECDH_PRIVATE_P256_MAGIC, magic.BCRYPT_ECDH_PRIVATE_P384_MAGIC, magic.BCRYPT_ECDH_PRIVATE_P521_MAGIC,
		magic.BCRYPT_ECDSA_PRIVATE_P256_MAGIC, magic.BCRYPT_ECDSA_PRIVATE_P384_MAGIC, magic.BCRYPT_ECDSA_PRIVATE_P521_MAGIC:
		blob := keys.BCRYPT_ECC_PRIVATE_KEY{}
		bytesRead, err := blob.Unmarshal(data)
		if err != nil {
			return nil, 0, err
		}
		return &blob, bytesRead, nil

	case magic.BCRYPT_DSA_PUBLIC_MAGIC:
		blob := keys.BCRYPT_DSA_PUBLIC_KEY{}
		bytesRead, err := blob.Unmarshal(data)
		if err != nil {
			return nil, 0, err
		}
		return &blob, bytesRead, nil

	case magic.BCRYPT_DSA_PRIVATE_MAGIC:
		blob := keys.BCRYPT_DSA_PRIVATE_KEY{}
		bytesRead, err := blob.Unmarshal(data)
		if err != nil {
			return nil, 0, err
		}
		return &blob, bytesRead, nil
	}

	return nil, 0, fmt.Errorf("invalid key material magic: 0x%08x", blobMagic.Magic)
}

package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION_INTERNAL_AES carries AES-encrypted trusted-domain
// authentication material ([MS-LSAD] 2.2.7.24). It has the same layout as
// LSAPR_AES_CIPHER_VALUE: fixed 64-byte AuthData and 16-byte Salt, then a
// [size_is(cbCipher)] [unique] pointer to the ciphertext.
type LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION_INTERNAL_AES struct {
	AuthData [64]uint8
	Salt     [16]uint8
	CbCipher ndr.DWORD
	Cipher   []uint8 `ndr:"unique,size_is=CbCipher"`
}

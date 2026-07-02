package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_AES_CIPHER_VALUE carries an AES-encrypted secret value ([MS-LSAD] 2.2.6.4).
// AuthData and Salt are fixed 64- and 16-byte fields; Cipher is a [size_is(cbCipher)]
// [unique] pointer to the ciphertext.
type LSAPR_AES_CIPHER_VALUE struct {
	AuthData [64]uint8
	Salt     [16]uint8
	CbCipher ndr.DWORD
	Cipher   []uint8 `ndr:"unique,size_is=CbCipher"`
}

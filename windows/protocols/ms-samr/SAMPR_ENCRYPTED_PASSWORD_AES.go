package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_ENCRYPTED_PASSWORD_AES carries an AES-encrypted password along with the
// authentication data, salt, and key-derivation parameters ([MS-SAMR]
// 2.2.6.32). Cipher is a [unique] pointer to a conformant array of CbCipher
// bytes.
type SAMPR_ENCRYPTED_PASSWORD_AES struct {
	AuthData         [64]byte
	Salt             [16]byte
	CbCipher         ndr.DWORD
	Cipher           []byte `ndr:"unique,size_is=CbCipher"`
	PBKDF2Iterations uint64
}

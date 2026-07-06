package mswkst

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// JOINPR_USER_PASSWORD is the cleartext join password structure ([MS-WKST] 2.2.5.17).
// All fields are fixed-size: Obfuscator is JOIN_OBFUSCATOR_LENGTH (8) bytes, Buffer is
// JOIN_MAX_PASSWORD_LENGTH (256) wchar_t, and Length is the byte length of the password
// stored at the tail of Buffer. The structure is never sent as plaintext on the wire; it
// is encrypted into a JOINPR_ENCRYPTED_USER_PASSWORD first. The arrays are FIXED NDR
// arrays (no conformance, no referent id), so they are modeled as Go fixed arrays.
type JOINPR_USER_PASSWORD struct {
	Obfuscator [8]uint8
	Buffer     [256]uint16
	Length     ndr.DWORD
}

package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAM_VALIDATE_PASSWORD_HASH holds the hash of a cleartext password
// ([MS-SAMR] 2.2.9.2). Hash is a [unique] pointer to a conformant array of
// Length bytes.
type SAM_VALIDATE_PASSWORD_HASH struct {
	Length ndr.DWORD
	Hash   []byte `ndr:"unique,size_is=Length"`
}

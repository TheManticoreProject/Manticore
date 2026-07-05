package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAM_VALIDATE_PERSISTED_FIELDS holds user account properties that the client is
// expected to persist across calls to SamrValidatePassword ([MS-SAMR] 2.2.9.3).
// PasswordHistory is a [unique] pointer to a conformant array of
// PasswordHistoryLength SAM_VALIDATE_PASSWORD_HASH structures.
type SAM_VALIDATE_PERSISTED_FIELDS struct {
	PresentFields         ndr.DWORD
	PasswordLastSet       dtyp.LARGE_INTEGER
	BadPasswordTime       dtyp.LARGE_INTEGER
	LockoutTime           dtyp.LARGE_INTEGER
	BadPasswordCount      ndr.DWORD
	PasswordHistoryLength ndr.DWORD
	PasswordHistory       []SAM_VALIDATE_PASSWORD_HASH `ndr:"unique,size_is=PasswordHistoryLength"`
}

package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAM_VALIDATE_PERSISTED_FIELDS holds user account properties that the client is
// expected to persist across calls to SamrValidatePassword ([MS-SAMR] 2.2.9.3).
// PasswordHistory is a [unique] pointer to a conformant array of
// PasswordHistoryLength SAM_VALIDATE_PASSWORD_HASH structures.
type SAM_VALIDATE_PERSISTED_FIELDS struct {
	PresentFields         ndr.DWORD
	PasswordLastSet       msdtyp.LARGE_INTEGER
	BadPasswordTime       msdtyp.LARGE_INTEGER
	LockoutTime           msdtyp.LARGE_INTEGER
	BadPasswordCount      ndr.DWORD
	PasswordHistoryLength ndr.DWORD
	PasswordHistory       []SAM_VALIDATE_PASSWORD_HASH `ndr:"unique,size_is=PasswordHistoryLength"`
}

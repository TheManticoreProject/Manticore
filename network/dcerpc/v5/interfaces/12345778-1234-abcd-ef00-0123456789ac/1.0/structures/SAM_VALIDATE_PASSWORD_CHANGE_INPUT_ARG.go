package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAM_VALIDATE_PASSWORD_CHANGE_INPUT_ARG holds the input for a
// SamValidatePasswordChange password validation request ([MS-SAMR] 2.2.9.7).
type SAM_VALIDATE_PASSWORD_CHANGE_INPUT_ARG struct {
	InputPersistedFields SAM_VALIDATE_PERSISTED_FIELDS
	ClearPassword        dtyp.RPC_UNICODE_STRING
	UserAccountName      dtyp.RPC_UNICODE_STRING
	HashedPassword       SAM_VALIDATE_PASSWORD_HASH
	PasswordMatch        uint8
}

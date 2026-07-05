package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAM_VALIDATE_PASSWORD_RESET_INPUT_ARG holds the input for a
// SamValidatePasswordReset password validation request ([MS-SAMR] 2.2.9.8).
type SAM_VALIDATE_PASSWORD_RESET_INPUT_ARG struct {
	InputPersistedFields          SAM_VALIDATE_PERSISTED_FIELDS
	ClearPassword                 dtyp.RPC_UNICODE_STRING
	UserAccountName               dtyp.RPC_UNICODE_STRING
	HashedPassword                SAM_VALIDATE_PASSWORD_HASH
	PasswordMustChangeAtNextLogon uint8
	ClearLockout                  uint8
}

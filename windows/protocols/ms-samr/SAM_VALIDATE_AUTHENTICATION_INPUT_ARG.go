package mssamr

// SAM_VALIDATE_AUTHENTICATION_INPUT_ARG holds the input for a
// SamValidateAuthentication password validation request ([MS-SAMR] 2.2.9.6).
type SAM_VALIDATE_AUTHENTICATION_INPUT_ARG struct {
	InputPersistedFields SAM_VALIDATE_PERSISTED_FIELDS
	PasswordMatched      uint8
}

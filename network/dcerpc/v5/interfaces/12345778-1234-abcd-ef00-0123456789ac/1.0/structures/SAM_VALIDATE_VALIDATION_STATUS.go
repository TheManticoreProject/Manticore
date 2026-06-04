package structures

// SAM_VALIDATE_VALIDATION_STATUS indicates the result of a password validation
// operation ([MS-SAMR] 2.2.9.4). As an NDR enum it is transmitted as a 16-bit
// unsigned value ([C706] section 14.3.6).
type SAM_VALIDATE_VALIDATION_STATUS uint16

const (
	SamValidateSuccess                  SAM_VALIDATE_VALIDATION_STATUS = 0
	SamValidatePasswordMustChange       SAM_VALIDATE_VALIDATION_STATUS = 1
	SamValidateAccountLockedOut         SAM_VALIDATE_VALIDATION_STATUS = 2
	SamValidatePasswordExpired          SAM_VALIDATE_VALIDATION_STATUS = 3
	SamValidatePasswordIncorrect        SAM_VALIDATE_VALIDATION_STATUS = 4
	SamValidatePasswordIsInHistory      SAM_VALIDATE_VALIDATION_STATUS = 5
	SamValidatePasswordTooShort         SAM_VALIDATE_VALIDATION_STATUS = 6
	SamValidatePasswordTooLong          SAM_VALIDATE_VALIDATION_STATUS = 7
	SamValidatePasswordNotComplexEnough SAM_VALIDATE_VALIDATION_STATUS = 8
	SamValidatePasswordTooRecent        SAM_VALIDATE_VALIDATION_STATUS = 9
	SamValidatePasswordFilterError      SAM_VALIDATE_VALIDATION_STATUS = 10
)

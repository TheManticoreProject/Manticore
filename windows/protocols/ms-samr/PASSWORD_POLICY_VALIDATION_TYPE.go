package mssamr

// PASSWORD_POLICY_VALIDATION_TYPE indicates the type of password validation
// being requested ([MS-SAMR] 2.2.9.1). As an NDR enum it is transmitted as a
// 16-bit unsigned value ([C706] section 14.3.6) and serves as the discriminant
// for the SAM_VALIDATE_INPUT_ARG and SAM_VALIDATE_OUTPUT_ARG unions.
type PASSWORD_POLICY_VALIDATION_TYPE uint16

const (
	SamValidateAuthentication PASSWORD_POLICY_VALIDATION_TYPE = 1
	SamValidatePasswordChange PASSWORD_POLICY_VALIDATION_TYPE = 2
	SamValidatePasswordReset  PASSWORD_POLICY_VALIDATION_TYPE = 3
)

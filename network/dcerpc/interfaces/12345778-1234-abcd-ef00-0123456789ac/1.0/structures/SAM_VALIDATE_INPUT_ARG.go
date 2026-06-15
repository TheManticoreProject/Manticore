package structures

// SAM_VALIDATE_INPUT_ARG is the discriminated union of password validation input
// arguments ([MS-SAMR] 2.2.9.9). The switch_type is PASSWORD_POLICY_VALIDATION_TYPE;
// the wire form is the discriminant followed by the single selected arm
// ([C706] section 14.3.8). Each arm is a value field for the corresponding
// validation type.
type SAM_VALIDATE_INPUT_ARG struct {
	Tag                         PASSWORD_POLICY_VALIDATION_TYPE        `ndr:"switch,enum"`
	ValidateAuthenticationInput SAM_VALIDATE_AUTHENTICATION_INPUT_ARG  `ndr:"case=1"`
	ValidatePasswordChangeInput SAM_VALIDATE_PASSWORD_CHANGE_INPUT_ARG `ndr:"case=2"`
	ValidatePasswordResetInput  SAM_VALIDATE_PASSWORD_RESET_INPUT_ARG  `ndr:"case=3"`
}

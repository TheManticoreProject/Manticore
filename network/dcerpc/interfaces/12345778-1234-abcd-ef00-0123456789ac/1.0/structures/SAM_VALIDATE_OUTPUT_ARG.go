package structures

// SAM_VALIDATE_OUTPUT_ARG is the discriminated union of password validation output
// arguments ([MS-SAMR] 2.2.9.10). The switch_type is PASSWORD_POLICY_VALIDATION_TYPE;
// the wire form is the discriminant followed by the single selected arm
// ([C706] section 14.3.8). All three arms carry a SAM_VALIDATE_STANDARD_OUTPUT_ARG.
type SAM_VALIDATE_OUTPUT_ARG struct {
	Tag                          PASSWORD_POLICY_VALIDATION_TYPE  `ndr:"switch,enum"`
	ValidateAuthenticationOutput SAM_VALIDATE_STANDARD_OUTPUT_ARG `ndr:"case=1"`
	ValidatePasswordChangeOutput SAM_VALIDATE_STANDARD_OUTPUT_ARG `ndr:"case=2"`
	ValidatePasswordResetOutput  SAM_VALIDATE_STANDARD_OUTPUT_ARG `ndr:"case=3"`
}

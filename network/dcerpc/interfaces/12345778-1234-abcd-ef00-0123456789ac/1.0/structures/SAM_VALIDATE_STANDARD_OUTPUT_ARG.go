package structures

// SAM_VALIDATE_STANDARD_OUTPUT_ARG is the common output of all
// SamrValidatePassword operations ([MS-SAMR] 2.2.9.5). It carries the persisted
// fields that the server has modified and the overall validation status.
type SAM_VALIDATE_STANDARD_OUTPUT_ARG struct {
	ChangedPersistedFields SAM_VALIDATE_PERSISTED_FIELDS
	ValidationStatus       SAM_VALIDATE_VALIDATION_STATUS `ndr:"enum"`
}

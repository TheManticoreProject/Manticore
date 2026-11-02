package status

import (
	"fmt"
	"strings"
)

// GSSAPIMajorStatus is a GSS-API major status (OM_uint32, RFC 2744 3.9.1). It
// packs three fields: a calling error in bits 24-31, a routine error in bits
// 16-23, and supplementary information bits in 0-15. GSS_S_COMPLETE (0) means
// success with no supplementary information.
type GSSAPIMajorStatus uint32

// GSSAPIMinorStatus is a mechanism-specific minor status, whose values are
// defined by the underlying security mechanism rather than by GSS-API.
type GSSAPIMinorStatus uint32

// Field positions and masks of a major status (RFC 2744 3.9.1).
const (
	callingErrorOffset = 24
	routineErrorOffset = 16

	callingErrorMask  GSSAPIMajorStatus = 0xff << callingErrorOffset
	routineErrorMask  GSSAPIMajorStatus = 0xff << routineErrorOffset
	supplementaryMask GSSAPIMajorStatus = 0xffff
)

// GSSAPI status codes, with the values of RFC 2744 3.9.1.
// Sources:
// https://www.rfc-editor.org/rfc/rfc2744.html#section-3.9.1
// https://www.rfc-editor.org/rfc/rfc2743.html#section-1.2.1.1
const (
	// GSS_S_COMPLETE is successful completion.
	GSS_S_COMPLETE GSSAPIMajorStatus = 0

	// CALLING ERRORS (bits 24-31)
	GSS_S_CALL_INACCESSIBLE_READ  GSSAPIMajorStatus = 1 << callingErrorOffset // a required input parameter could not be read
	GSS_S_CALL_INACCESSIBLE_WRITE GSSAPIMajorStatus = 2 << callingErrorOffset // a required output parameter could not be written
	GSS_S_CALL_BAD_STRUCTURE      GSSAPIMajorStatus = 3 << callingErrorOffset // a parameter was malformed

	// ROUTINE ERRORS (bits 16-23)
	GSS_S_BAD_MECH             GSSAPIMajorStatus = 1 << routineErrorOffset  // unsupported mechanism requested
	GSS_S_BAD_NAME             GSSAPIMajorStatus = 2 << routineErrorOffset  // invalid name provided
	GSS_S_BAD_NAMETYPE         GSSAPIMajorStatus = 3 << routineErrorOffset  // name of unsupported type provided
	GSS_S_BAD_BINDINGS         GSSAPIMajorStatus = 4 << routineErrorOffset  // channel binding mismatch
	GSS_S_BAD_STATUS           GSSAPIMajorStatus = 5 << routineErrorOffset  // invalid input status selector
	GSS_S_BAD_SIG              GSSAPIMajorStatus = 6 << routineErrorOffset  // token had invalid integrity check
	GSS_S_BAD_MIC              GSSAPIMajorStatus = GSS_S_BAD_SIG            // preferred alias for GSS_S_BAD_SIG
	GSS_S_NO_CRED              GSSAPIMajorStatus = 7 << routineErrorOffset  // no valid credentials provided
	GSS_S_NO_CONTEXT           GSSAPIMajorStatus = 8 << routineErrorOffset  // no valid security context specified
	GSS_S_DEFECTIVE_TOKEN      GSSAPIMajorStatus = 9 << routineErrorOffset  // defective token detected
	GSS_S_DEFECTIVE_CREDENTIAL GSSAPIMajorStatus = 10 << routineErrorOffset // defective credential detected
	GSS_S_CREDENTIALS_EXPIRED  GSSAPIMajorStatus = 11 << routineErrorOffset // expired credentials detected
	GSS_S_CONTEXT_EXPIRED      GSSAPIMajorStatus = 12 << routineErrorOffset // specified security context expired
	GSS_S_FAILURE              GSSAPIMajorStatus = 13 << routineErrorOffset // failure, unspecified at GSS-API level
	GSS_S_BAD_QOP              GSSAPIMajorStatus = 14 << routineErrorOffset // unsupported QOP value
	GSS_S_UNAUTHORIZED         GSSAPIMajorStatus = 15 << routineErrorOffset // operation unauthorized
	GSS_S_UNAVAILABLE          GSSAPIMajorStatus = 16 << routineErrorOffset // operation unavailable
	GSS_S_DUPLICATE_ELEMENT    GSSAPIMajorStatus = 17 << routineErrorOffset // duplicate credential element requested
	GSS_S_NAME_NOT_MN          GSSAPIMajorStatus = 18 << routineErrorOffset // name contains multi-mechanism elements

	// SUPPLEMENTARY INFORMATION BITS (bits 0-15)
	GSS_S_CONTINUE_NEEDED GSSAPIMajorStatus = 1 << 0 // continuation call to routine required; returned only by gss_init_sec_context() or gss_accept_sec_context()
	GSS_S_DUPLICATE_TOKEN GSSAPIMajorStatus = 1 << 1 // duplicate per-message token detected
	GSS_S_OLD_TOKEN       GSSAPIMajorStatus = 1 << 2 // timed-out per-message token detected
	GSS_S_UNSEQ_TOKEN     GSSAPIMajorStatus = 1 << 3 // reordered (early) per-message token detected
	GSS_S_GAP_TOKEN       GSSAPIMajorStatus = 1 << 4 // skipped predecessor token(s) detected
)

// GSSAPIMajorStatusMessages describes each status code, keyed by the code's own
// value (a single calling error, routine error or supplementary bit).
var GSSAPIMajorStatusMessages = map[GSSAPIMajorStatus]string{
	GSS_S_COMPLETE: "Successful completion",

	// CALLING ERRORS
	GSS_S_CALL_INACCESSIBLE_READ:  "A required input parameter could not be read",
	GSS_S_CALL_INACCESSIBLE_WRITE: "A required output parameter could not be written",
	GSS_S_CALL_BAD_STRUCTURE:      "A parameter was malformed",

	// ROUTINE ERRORS
	GSS_S_BAD_MECH:             "Unsupported mechanism requested",
	GSS_S_BAD_NAME:             "Invalid name provided",
	GSS_S_BAD_NAMETYPE:         "Name of unsupported type provided",
	GSS_S_BAD_BINDINGS:         "Channel binding mismatch",
	GSS_S_BAD_STATUS:           "Invalid input status selector",
	GSS_S_BAD_MIC:              "Token had invalid integrity check",
	GSS_S_NO_CRED:              "No valid credentials provided",
	GSS_S_NO_CONTEXT:           "No valid security context specified",
	GSS_S_DEFECTIVE_TOKEN:      "Defective token detected",
	GSS_S_DEFECTIVE_CREDENTIAL: "Defective credential detected",
	GSS_S_CREDENTIALS_EXPIRED:  "Expired credentials detected",
	GSS_S_CONTEXT_EXPIRED:      "Specified security context expired",
	GSS_S_FAILURE:              "Failure, unspecified at GSS-API level",
	GSS_S_BAD_QOP:              "Unsupported QOP value",
	GSS_S_UNAUTHORIZED:         "Operation unauthorized",
	GSS_S_UNAVAILABLE:          "Operation unavailable",
	GSS_S_DUPLICATE_ELEMENT:    "Duplicate credential element requested",
	GSS_S_NAME_NOT_MN:          "Name contains multi-mechanism elements",

	// SUPPLEMENTARY INFORMATION
	GSS_S_CONTINUE_NEEDED: "Continuation call to routine required",
	GSS_S_DUPLICATE_TOKEN: "Duplicate per-message token detected",
	GSS_S_OLD_TOKEN:       "Timed-out per-message token detected",
	GSS_S_UNSEQ_TOKEN:     "Reordered (early) per-message token detected",
	GSS_S_GAP_TOKEN:       "Skipped predecessor token(s) detected",
}

// codeNames names each status code, keyed by its value.
var codeNames = map[GSSAPIMajorStatus]string{
	GSS_S_CALL_INACCESSIBLE_READ:  "GSS_S_CALL_INACCESSIBLE_READ",
	GSS_S_CALL_INACCESSIBLE_WRITE: "GSS_S_CALL_INACCESSIBLE_WRITE",
	GSS_S_CALL_BAD_STRUCTURE:      "GSS_S_CALL_BAD_STRUCTURE",
	GSS_S_BAD_MECH:                "GSS_S_BAD_MECH",
	GSS_S_BAD_NAME:                "GSS_S_BAD_NAME",
	GSS_S_BAD_NAMETYPE:            "GSS_S_BAD_NAMETYPE",
	GSS_S_BAD_BINDINGS:            "GSS_S_BAD_BINDINGS",
	GSS_S_BAD_STATUS:              "GSS_S_BAD_STATUS",
	GSS_S_BAD_MIC:                 "GSS_S_BAD_MIC",
	GSS_S_NO_CRED:                 "GSS_S_NO_CRED",
	GSS_S_NO_CONTEXT:              "GSS_S_NO_CONTEXT",
	GSS_S_DEFECTIVE_TOKEN:         "GSS_S_DEFECTIVE_TOKEN",
	GSS_S_DEFECTIVE_CREDENTIAL:    "GSS_S_DEFECTIVE_CREDENTIAL",
	GSS_S_CREDENTIALS_EXPIRED:     "GSS_S_CREDENTIALS_EXPIRED",
	GSS_S_CONTEXT_EXPIRED:         "GSS_S_CONTEXT_EXPIRED",
	GSS_S_FAILURE:                 "GSS_S_FAILURE",
	GSS_S_BAD_QOP:                 "GSS_S_BAD_QOP",
	GSS_S_UNAUTHORIZED:            "GSS_S_UNAUTHORIZED",
	GSS_S_UNAVAILABLE:             "GSS_S_UNAVAILABLE",
	GSS_S_DUPLICATE_ELEMENT:       "GSS_S_DUPLICATE_ELEMENT",
	GSS_S_NAME_NOT_MN:             "GSS_S_NAME_NOT_MN",
	GSS_S_CONTINUE_NEEDED:         "GSS_S_CONTINUE_NEEDED",
	GSS_S_DUPLICATE_TOKEN:         "GSS_S_DUPLICATE_TOKEN",
	GSS_S_OLD_TOKEN:               "GSS_S_OLD_TOKEN",
	GSS_S_UNSEQ_TOKEN:             "GSS_S_UNSEQ_TOKEN",
	GSS_S_GAP_TOKEN:               "GSS_S_GAP_TOKEN",
}

// CallingError returns the calling-error field of the status (GSS_CALLING_ERROR).
func (code GSSAPIMajorStatus) CallingError() GSSAPIMajorStatus { return code & callingErrorMask }

// RoutineError returns the routine-error field of the status (GSS_ROUTINE_ERROR).
func (code GSSAPIMajorStatus) RoutineError() GSSAPIMajorStatus { return code & routineErrorMask }

// SupplementaryInfo returns the supplementary-information bits of the status
// (GSS_SUPPLEMENTARY_INFO).
func (code GSSAPIMajorStatus) SupplementaryInfo() GSSAPIMajorStatus {
	return code & supplementaryMask
}

// IsError reports whether the status carries a calling or routine error
// (GSS_ERROR). Supplementary bits alone, such as GSS_S_CONTINUE_NEEDED, are not
// errors.
func (code GSSAPIMajorStatus) IsError() bool {
	return code&(callingErrorMask|routineErrorMask) != 0
}

// String names the status: "GSS_S_COMPLETE" for 0, otherwise the calling error,
// routine error and supplementary bits it carries joined with "|". A field value
// with no defined name is shown in hexadecimal.
func (code GSSAPIMajorStatus) String() string {
	if code == GSS_S_COMPLETE {
		return "GSS_S_COMPLETE"
	}
	var parts []string
	name := func(v GSSAPIMajorStatus) string {
		if n, ok := codeNames[v]; ok {
			return n
		}
		return fmt.Sprintf("UNKNOWN_GSS_STATUS(0x%08x)", uint32(v))
	}
	if v := code.CallingError(); v != 0 {
		parts = append(parts, name(v))
	}
	if v := code.RoutineError(); v != 0 {
		parts = append(parts, name(v))
	}
	for bit := GSSAPIMajorStatus(1); bit <= supplementaryMask && bit != 0; bit <<= 1 {
		if code&bit != 0 {
			parts = append(parts, name(bit))
		}
	}
	return strings.Join(parts, "|")
}

// String renders a mechanism-specific minor status in hexadecimal; its meaning is
// defined by the mechanism.
func (code GSSAPIMinorStatus) String() string {
	return fmt.Sprintf("0x%08x", uint32(code))
}

package status_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/gssapi/status"
)

// TestGSSAPIStatusValuesMatchRFC2744 pins the numeric values of RFC 2744 3.9.1:
// routine errors in bits 16-23, calling errors in 24-31, supplementary bits in
// 0-15, and GSS_S_COMPLETE equal to 0 and distinct from GSS_S_CONTINUE_NEEDED.
func TestGSSAPIStatusValuesMatchRFC2744(t *testing.T) {
	for _, tc := range []struct {
		code status.GSSAPIMajorStatus
		want uint32
	}{
		{status.GSS_S_COMPLETE, 0},
		{status.GSS_S_CALL_INACCESSIBLE_READ, 0x01000000},
		{status.GSS_S_CALL_INACCESSIBLE_WRITE, 0x02000000},
		{status.GSS_S_CALL_BAD_STRUCTURE, 0x03000000},
		{status.GSS_S_BAD_MECH, 0x00010000},
		{status.GSS_S_BAD_SIG, 0x00060000},
		{status.GSS_S_BAD_MIC, 0x00060000},
		{status.GSS_S_FAILURE, 0x000d0000},
		{status.GSS_S_NAME_NOT_MN, 0x00120000},
		{status.GSS_S_CONTINUE_NEEDED, 0x00000001},
		{status.GSS_S_DUPLICATE_TOKEN, 0x00000002},
		{status.GSS_S_OLD_TOKEN, 0x00000004},
		{status.GSS_S_UNSEQ_TOKEN, 0x00000008},
		{status.GSS_S_GAP_TOKEN, 0x00000010},
	} {
		if uint32(tc.code) != tc.want {
			t.Errorf("%s = 0x%08x, want 0x%08x", tc.code, uint32(tc.code), tc.want)
		}
	}
}

func TestGSSAPIMajorStatusString(t *testing.T) {
	tests := []struct {
		code     status.GSSAPIMajorStatus
		expected string
	}{
		{status.GSS_S_COMPLETE, "GSS_S_COMPLETE"},
		{status.GSS_S_BAD_MECH, "GSS_S_BAD_MECH"},
		{status.GSS_S_BAD_NAME, "GSS_S_BAD_NAME"},
		{status.GSS_S_BAD_NAMETYPE, "GSS_S_BAD_NAMETYPE"},
		{status.GSS_S_BAD_BINDINGS, "GSS_S_BAD_BINDINGS"},
		{status.GSS_S_BAD_STATUS, "GSS_S_BAD_STATUS"},
		{status.GSS_S_BAD_MIC, "GSS_S_BAD_MIC"},
		{status.GSS_S_NO_CRED, "GSS_S_NO_CRED"},
		{status.GSS_S_NO_CONTEXT, "GSS_S_NO_CONTEXT"},
		{status.GSS_S_DEFECTIVE_TOKEN, "GSS_S_DEFECTIVE_TOKEN"},
		{status.GSS_S_DEFECTIVE_CREDENTIAL, "GSS_S_DEFECTIVE_CREDENTIAL"},
		{status.GSS_S_CREDENTIALS_EXPIRED, "GSS_S_CREDENTIALS_EXPIRED"},
		{status.GSS_S_CONTEXT_EXPIRED, "GSS_S_CONTEXT_EXPIRED"},
		{status.GSS_S_FAILURE, "GSS_S_FAILURE"},
		{status.GSS_S_BAD_QOP, "GSS_S_BAD_QOP"},
		{status.GSS_S_UNAUTHORIZED, "GSS_S_UNAUTHORIZED"},
		{status.GSS_S_UNAVAILABLE, "GSS_S_UNAVAILABLE"},
		{status.GSS_S_DUPLICATE_ELEMENT, "GSS_S_DUPLICATE_ELEMENT"},
		{status.GSS_S_NAME_NOT_MN, "GSS_S_NAME_NOT_MN"},
		{status.GSS_S_CALL_BAD_STRUCTURE, "GSS_S_CALL_BAD_STRUCTURE"},
		{status.GSS_S_CONTINUE_NEEDED, "GSS_S_CONTINUE_NEEDED"},
		{status.GSS_S_GAP_TOKEN, "GSS_S_GAP_TOKEN"},
		{status.GSS_S_CALL_INACCESSIBLE_READ | status.GSS_S_FAILURE | status.GSS_S_OLD_TOKEN, "GSS_S_CALL_INACCESSIBLE_READ|GSS_S_FAILURE|GSS_S_OLD_TOKEN"},
		{status.GSSAPIMajorStatus(99 << 16), "UNKNOWN_GSS_STATUS(0x00630000)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.code.String(); got != tt.expected {
				t.Errorf("GSSAPIMajorStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestGSSAPIMajorStatusFields checks the GSS_CALLING_ERROR / GSS_ROUTINE_ERROR /
// GSS_SUPPLEMENTARY_INFO / GSS_ERROR accessors.
func TestGSSAPIMajorStatusFields(t *testing.T) {
	code := status.GSS_S_CALL_BAD_STRUCTURE | status.GSS_S_NO_CRED | status.GSS_S_DUPLICATE_TOKEN
	if code.CallingError() != status.GSS_S_CALL_BAD_STRUCTURE ||
		code.RoutineError() != status.GSS_S_NO_CRED ||
		code.SupplementaryInfo() != status.GSS_S_DUPLICATE_TOKEN {
		t.Errorf("fields of %s: %s / %s / %s", code, code.CallingError(), code.RoutineError(), code.SupplementaryInfo())
	}
	if !code.IsError() || !status.GSS_S_FAILURE.IsError() {
		t.Error("an error status is not reported as an error")
	}
	if status.GSS_S_COMPLETE.IsError() || status.GSS_S_CONTINUE_NEEDED.IsError() {
		t.Error("success or GSS_S_CONTINUE_NEEDED reported as an error")
	}
}

func TestGSSAPIMajorStatusMessages(t *testing.T) {
	tests := []struct {
		code     status.GSSAPIMajorStatus
		expected string
	}{
		{status.GSS_S_COMPLETE, "Successful completion"},
		{status.GSS_S_BAD_MECH, "Unsupported mechanism requested"},
		{status.GSS_S_FAILURE, "Failure, unspecified at GSS-API level"},
		{status.GSS_S_NAME_NOT_MN, "Name contains multi-mechanism elements"},
		{status.GSS_S_CONTINUE_NEEDED, "Continuation call to routine required"},
		{status.GSS_S_GAP_TOKEN, "Skipped predecessor token(s) detected"},
		{status.GSS_S_CALL_INACCESSIBLE_READ, "A required input parameter could not be read"},
	}

	for _, tt := range tests {
		t.Run(tt.code.String(), func(t *testing.T) {
			if got := status.GSSAPIMajorStatusMessages[tt.code]; got != tt.expected {
				t.Errorf("GSSAPIMajorStatusMessages[%v] = %v, want %v", tt.code, got, tt.expected)
			}
		})
	}
}

func TestGSSAPIMinorStatusString(t *testing.T) {
	if got := status.GSSAPIMinorStatus(0x96c73a23).String(); got != "0x96c73a23" {
		t.Errorf("GSSAPIMinorStatus.String() = %q", got)
	}
}

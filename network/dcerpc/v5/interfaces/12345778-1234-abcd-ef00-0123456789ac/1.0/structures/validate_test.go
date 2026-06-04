package structures

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTripValidate marshals in, unmarshals into a fresh value of the same type,
// and asserts deep equality. It is named distinctly to avoid colliding with any
// shared roundTrip helper other structure tests may define in this package.
func roundTripValidate[T any](t *testing.T, name string, in T) {
	t.Helper()
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("%s: Marshal: %v", name, err)
	}
	var out T
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: Unmarshal: %v", name, err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("%s: round trip mismatch:\n in:  %+v\n out: %+v", name, in, out)
	}
}

// TestSAM_VALIDATE_OUTPUT_ARG_RoundTrip exercises the SAM_VALIDATE_OUTPUT_ARG
// union with the SamValidateAuthentication arm (case 1) selected, including a
// SAM_VALIDATE_PERSISTED_FIELDS payload with a [unique] pointer to a conformant
// array of SAM_VALIDATE_PASSWORD_HASH (itself a [unique] byte buffer).
func TestSAM_VALIDATE_OUTPUT_ARG_RoundTrip(t *testing.T) {
	roundTripValidate(t, "ValidateAuthenticationOutput(case1)", SAM_VALIDATE_OUTPUT_ARG{
		Tag: SamValidateAuthentication,
		ValidateAuthenticationOutput: SAM_VALIDATE_STANDARD_OUTPUT_ARG{
			ChangedPersistedFields: SAM_VALIDATE_PERSISTED_FIELDS{
				PresentFields:         0x00000007,
				PasswordLastSet:       dtyp.LARGE_INTEGER(0x1122334455667788),
				BadPasswordTime:       dtyp.LARGE_INTEGER(0x0000000000000000),
				LockoutTime:           dtyp.LARGE_INTEGER(0x00000000DEADBEEF),
				BadPasswordCount:      3,
				PasswordHistoryLength: 2,
				PasswordHistory: []SAM_VALIDATE_PASSWORD_HASH{
					{Length: 4, Hash: []byte{0x01, 0x02, 0x03, 0x04}},
					{Length: 2, Hash: []byte{0xAA, 0xBB}},
				},
			},
			ValidationStatus: SamValidatePasswordExpired,
		},
	})
}

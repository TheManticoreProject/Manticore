package structures

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestSamprUserInfoBufferNameRoundTrip marshals a SAMPR_USER_INFO_BUFFER whose
// discriminant selects UserNameInformation (6) and verifies it survives an NDR
// round trip unchanged.
func TestSamprUserInfoBufferNameRoundTrip(t *testing.T) {
	in := SAMPR_USER_INFO_BUFFER{
		Tag: UserNameInformation,
		Name: SAMPR_USER_NAME_INFORMATION{
			UserName: dtyp.NewUnicodeString("administrator"),
			FullName: dtyp.NewUnicodeString("Administrator Account"),
		},
	}

	data, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var out SAMPR_USER_INFO_BUFFER
	if err := ndr.Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round trip mismatch:\n in: %+v\nout: %+v", in, out)
	}
}

// TestSamprUserInfoBufferControlRoundTrip marshals a SAMPR_USER_INFO_BUFFER
// whose discriminant selects UserControlInformation (16) and verifies it
// survives an NDR round trip unchanged.
func TestSamprUserInfoBufferControlRoundTrip(t *testing.T) {
	in := SAMPR_USER_INFO_BUFFER{
		Tag: UserControlInformation,
		Control: USER_CONTROL_INFORMATION{
			UserAccountControl: 0x00000200,
		},
	}

	data, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var out SAMPR_USER_INFO_BUFFER
	if err := ndr.Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round trip mismatch:\n in: %+v\nout: %+v", in, out)
	}
}

package ldap

import (
	"testing"

	ber "github.com/go-asn1-ber/asn1-ber"
)

// TestSecurityInformationDefaultExcludesSACL pins the value an unprivileged read has
// to ask for. Letting SACL_SECURITY_INFORMATION into this set makes a domain
// controller return the nTSecurityDescriptor attribute empty, with no error, for every
// client that does not hold SE_SECURITY_NAME.
func TestSecurityInformationDefaultExcludesSACL(t *testing.T) {
	if SECURITY_INFORMATION_DEFAULT != 0x7 {
		t.Errorf("SECURITY_INFORMATION_DEFAULT = 0x%x, want 0x7", SECURITY_INFORMATION_DEFAULT)
	}
	if SECURITY_INFORMATION_DEFAULT&SACL_SECURITY_INFORMATION != 0 {
		t.Error("SECURITY_INFORMATION_DEFAULT must not request the SACL")
	}
}

// TestControlMicrosoftSDFlagsGetControlType checks the OID the control is sent under.
func TestControlMicrosoftSDFlagsGetControlType(t *testing.T) {
	control := NewControlMicrosoftSDFlags()
	if got, want := control.GetControlType(), "1.2.840.113556.1.4.801"; got != want {
		t.Errorf("GetControlType() = %q, want %q", got, want)
	}
}

// TestControlMicrosoftSDFlagsEncode walks the encoded control and checks that the OID
// and the flags value are where a server expects to read them: a SEQUENCE holding the
// control OID and an OCTET STRING wrapping a SEQUENCE with a single INTEGER.
func TestControlMicrosoftSDFlagsEncode(t *testing.T) {
	testCases := []struct {
		name         string
		criticality  bool
		controlValue int32
		wantChildren int
	}{
		{"default flags, not critical", false, SECURITY_INFORMATION_DEFAULT, 2},
		{"with SACL, not critical", false, SECURITY_INFORMATION_DEFAULT | SACL_SECURITY_INFORMATION, 2},
		{"default flags, critical", true, SECURITY_INFORMATION_DEFAULT, 3},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			control := &ControlMicrosoftSDFlags{Criticality: testCase.criticality, ControlValue: testCase.controlValue}
			packet := control.Encode()

			if len(packet.Children) != testCase.wantChildren {
				t.Fatalf("encoded control has %d children, want %d", len(packet.Children), testCase.wantChildren)
			}

			oid, ok := packet.Children[0].Value.(string)
			if !ok {
				t.Fatalf("first child is %T, want the control OID as a string", packet.Children[0].Value)
			}
			if oid != "1.2.840.113556.1.4.801" {
				t.Errorf("encoded OID = %q, want %q", oid, "1.2.840.113556.1.4.801")
			}

			// The control value is the last child whatever the criticality, since the
			// criticality boolean is only emitted when it is true.
			value := packet.Children[len(packet.Children)-1]
			if len(value.Children) != 1 {
				t.Fatalf("control value has %d children, want 1 SDFlags sequence", len(value.Children))
			}
			flags := value.Children[0]
			if len(flags.Children) != 1 {
				t.Fatalf("SDFlags sequence has %d children, want 1 integer", len(flags.Children))
			}
			if flags.Children[0].Tag != ber.TagInteger {
				t.Errorf("SDFlags child tag = %d, want TagInteger (%d)", flags.Children[0].Tag, ber.TagInteger)
			}
			// ber.NewInteger stores the value it was handed, so this is the int32 the
			// control carries rather than a widened int64.
			got, ok := flags.Children[0].Value.(int32)
			if !ok {
				t.Fatalf("SDFlags value is %T, want int32", flags.Children[0].Value)
			}
			if got != testCase.controlValue {
				t.Errorf("SDFlags value = %d, want %d", got, testCase.controlValue)
			}
		})
	}
}

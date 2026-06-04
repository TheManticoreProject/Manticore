package structures

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestSAMPR_DOMAIN_INFO_BUFFER_RoundTrip marshals and unmarshals a
// SAMPR_DOMAIN_INFO_BUFFER selecting the DomainNameInformation (5) arm and checks the
// recovered value matches the original.
func TestSAMPR_DOMAIN_INFO_BUFFER_RoundTrip(t *testing.T) {
	in := SAMPR_DOMAIN_INFO_BUFFER{
		Tag: DomainNameInformation,
		Name: SAMPR_DOMAIN_NAME_INFORMATION{
			DomainName: dtyp.NewUnicodeString("CONTOSO"),
		},
	}

	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var out SAMPR_DOMAIN_INFO_BUFFER
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if out.Tag != DomainNameInformation {
		t.Fatalf("Tag = %d, want %d", out.Tag, DomainNameInformation)
	}
	if got := out.Name.DomainName.String(); got != "CONTOSO" {
		t.Fatalf("DomainName = %q, want %q", got, "CONTOSO")
	}
}

// TestSAMPR_DISPLAY_INFO_BUFFER_RoundTrip marshals and unmarshals a
// SAMPR_DISPLAY_INFO_BUFFER selecting the DomainDisplayUser (1) arm with an empty user
// buffer and checks the recovered value matches the original.
func TestSAMPR_DISPLAY_INFO_BUFFER_RoundTrip(t *testing.T) {
	in := SAMPR_DISPLAY_INFO_BUFFER{
		Tag: DomainDisplayUser,
		UserInformation: SAMPR_DOMAIN_DISPLAY_USER_BUFFER{
			EntriesRead: 0,
			Buffer:      nil,
		},
	}

	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var out SAMPR_DISPLAY_INFO_BUFFER
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if out.Tag != DomainDisplayUser {
		t.Fatalf("Tag = %d, want %d", out.Tag, DomainDisplayUser)
	}
	if out.UserInformation.EntriesRead != 0 {
		t.Fatalf("EntriesRead = %d, want 0", out.UserInformation.EntriesRead)
	}
	if len(out.UserInformation.Buffer) != 0 {
		t.Fatalf("Buffer len = %d, want 0", len(out.UserInformation.Buffer))
	}
	_ = reflect.DeepEqual(in, out)
}

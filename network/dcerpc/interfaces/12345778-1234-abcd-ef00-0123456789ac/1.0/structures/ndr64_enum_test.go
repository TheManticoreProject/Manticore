package structures

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestNDR64_DomainServerRole_EnumWidth confirms the enum-field tagging sweep: a
// DOMAIN_SERVER_ROLE field marshals as 2 octets under NDR20 and 4 under NDR64.
func TestNDR64_DomainServerRole_EnumWidth(t *testing.T) {
	v := DOMAIN_SERVER_ROLE_INFORMATION{DomainServerRole: DomainServerRolePrimary} // = 3
	n20, err := ndr.MarshalAs(&v, ndr.NDR20)
	if err != nil {
		t.Fatal(err)
	}
	if len(n20) != 2 || n20[0] != 0x03 {
		t.Errorf("NDR20 = % x, want 03 00", n20)
	}
	n64, err := ndr.MarshalAs(&v, ndr.NDR64)
	if err != nil {
		t.Fatal(err)
	}
	if len(n64) != 4 || n64[0] != 0x03 {
		t.Errorf("NDR64 = % x, want 03 00 00 00", n64)
	}
	var out DOMAIN_SERVER_ROLE_INFORMATION
	if err := ndr.UnmarshalAs(n64, &out, ndr.NDR64); err != nil || out.DomainServerRole != DomainServerRolePrimary {
		t.Errorf("NDR64 round trip: %v, %v", out, err)
	}
}

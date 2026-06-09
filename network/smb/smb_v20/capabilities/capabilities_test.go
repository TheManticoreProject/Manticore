package capabilities_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/capabilities"
)

func TestCapabilitiesHas(t *testing.T) {
	c := capabilities.SMB2_GLOBAL_CAP_DFS | capabilities.SMB2_GLOBAL_CAP_LARGE_MTU
	if !c.Has(capabilities.SMB2_GLOBAL_CAP_DFS) {
		t.Errorf("expected DFS capability to be set")
	}
	if !c.Has(capabilities.SMB2_GLOBAL_CAP_DFS | capabilities.SMB2_GLOBAL_CAP_LARGE_MTU) {
		t.Errorf("expected combined capabilities to be set")
	}
	if c.Has(capabilities.SMB2_GLOBAL_CAP_ENCRYPTION) {
		t.Errorf("did not expect ENCRYPTION capability to be set")
	}
}

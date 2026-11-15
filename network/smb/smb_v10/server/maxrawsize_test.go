package server

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
)

// TestMaxRawSizeMatchesWindowsWithoutPromisingRawMode checks the NEGOTIATE
// MaxRawSize matches the Windows value (65536) while CAP_RAW_MODE stays
// unadvertised. The field is only meaningful once CAP_RAW_MODE is negotiated
// ([MS-CIFS] 2.2.4.52.2), so reporting the Windows value is a fingerprint fix
// that makes no behavioural promise. If anyone advertises CAP_RAW_MODE without
// implementing raw transfers, this test fails, because then the reported size
// would become a promise the server does not keep.
func TestMaxRawSizeMatchesWindowsWithoutPromisingRawMode(t *testing.T) {
	if defaultMaxRawSize != 65536 {
		t.Errorf("defaultMaxRawSize = %d, want 65536 (the Windows value)", defaultMaxRawSize)
	}
	if serverCapabilities&capabilities.CAP_RAW_MODE != 0 {
		t.Error("CAP_RAW_MODE is advertised but raw transfers are not implemented; " +
			"MaxRawSize must not be advertised as a real capability")
	}
}

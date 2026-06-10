package targetinfo_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
)

// TestBuildBlobTargetInfoRejectsOversizedAVPairLength verifies that an AVPair
// whose declared length runs past the end of the buffer does not panic.
func TestBuildBlobTargetInfoRejectsOversizedAVPairLength(t *testing.T) {
	// One AVPair header: AvId=MsvAvNbComputerName(0x0001), AvLen=0xFFFF, but no
	// value bytes follow — the declared length runs far past the buffer.
	malformed := []byte{0x01, 0x00, 0xFF, 0xFF}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("BuildBlobTargetInfo panicked on an oversized AVPair length: %v", r)
		}
	}()

	_ = targetinfo.BuildBlobTargetInfo(malformed)
}

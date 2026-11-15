package rpcpipe

import (
	"fmt"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server"
)

// TestDefaultVersionMatchesNativeOS pins the default wkssvc version to the SMB1
// server's default NativeOS/NativeLanMan identity. The two are reported on the
// same host, so a wkssvc version (formerly 6.1) that disagreed with the NativeOS
// string (Windows Server 2012 R2, i.e. 6.3) was a server fingerprint. If the
// default OS identity is ever changed, this test fails so the wkssvc default is
// updated to match.
func TestDefaultVersionMatchesNativeOS(t *testing.T) {
	want := fmt.Sprintf("%d.%d", defaultVersionMajor, defaultVersionMinor)
	if want != "6.3" {
		t.Fatalf("default wkssvc version = %s, want 6.3 to match the Server 2012 R2 NativeOS", want)
	}
	if !strings.Contains(server.DefaultNativeLanMan, want) {
		t.Fatalf("wkssvc default version %s does not appear in NativeLanMan %q", want, server.DefaultNativeLanMan)
	}
}

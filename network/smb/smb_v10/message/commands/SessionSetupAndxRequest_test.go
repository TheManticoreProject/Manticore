package commands_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestSessionSetupAndxRequest_ExtendedSecurityRoundTrip verifies that, in the
// extended-security variant, the SecurityBlob is preserved across a
// Marshal/Unmarshal cycle and that NativeOS/NativeLanMan are parsed from the
// correct offset (immediately after the blob). Per [MS-SMB] 2.2.4.5 / [MS-CIFS]
// 2.2.4.53.1, the data block is SecurityBlob[SecurityBlobLength] followed by
// NativeOS and NativeLanMan; SecurityBlobLength lives only in the parameters
// block, not inside the data block.
func TestSessionSetupAndxRequest_ExtendedSecurityRoundTrip(t *testing.T) {
	securityBlob := []types.UCHAR{0x60, 0x12, 0x06, 0x06, 0x2b, 0x06, 0x01, 0x05, 0x05, 0x02}

	req := commands.NewSessionSetupAndxRequest()
	req.MaxBufferSize = 0x4104
	req.MaxMpxCount = 50
	req.VcNumber = 0
	req.SessionKey = 0
	req.Capabilities = capabilities.CAP_EXTENDED_SECURITY
	req.SecurityBlob = securityBlob
	req.NativeOS = "Windows"
	req.NativeLanMan = "Manticore"

	marshalled, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	parsed := commands.NewSessionSetupAndxRequest()
	parsed.SetParameters(parameters.NewParameters())
	parsed.SetData(data.NewData())
	if _, err := parsed.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if parsed.SecurityBlobLength != types.USHORT(len(securityBlob)) {
		t.Errorf("SecurityBlobLength: expected %d, got %d", len(securityBlob), parsed.SecurityBlobLength)
	}

	if !bytes.Equal([]byte(parsed.SecurityBlob), []byte(securityBlob)) {
		t.Errorf("SecurityBlob: expected % x, got % x",
			[]byte(securityBlob), []byte(parsed.SecurityBlob))
	}

	if parsed.NativeOS != req.NativeOS {
		t.Errorf("NativeOS: expected %q, got %q", req.NativeOS, parsed.NativeOS)
	}

	if parsed.NativeLanMan != req.NativeLanMan {
		t.Errorf("NativeLanMan: expected %q, got %q", req.NativeLanMan, parsed.NativeLanMan)
	}
}

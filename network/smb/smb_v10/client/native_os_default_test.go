package client_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// TestSessionSetupDefaultsNativeStrings guards the fix for the empty-NativeOS bug:
// strict servers (Windows Server 2016) reject a SESSION_SETUP_ANDX whose
// NativeOS/NativeLanMan are empty, returning an empty challenge blob that then fails
// SPNEGO parsing with "asn1: sequence truncated". When the caller leaves those fields
// blank, SessionSetup must substitute non-empty defaults.
func TestSessionSetupDefaultsNativeStrings(t *testing.T) {
	// Share-level setup completes in a single round trip, so one canned response with a
	// reply UID is enough to drive SessionSetup to completion.
	resp := message.NewMessage()
	resp.Header.SetFlags(flags.FLAGS_REPLY)
	resp.Header.UID = 0x0901
	resp.AddCommand(commands.NewSessionSetupAndxResponse())
	raw, err := resp.Marshal()
	if err != nil {
		t.Fatalf("marshal canned response: %v", err)
	}

	tr := &capturingTransport{response: raw}
	c := &client.Client{
		Transport:  tr,
		Connection: &client.Connection{Server: &client.Server{SecurityMode: 0}},
	}
	// NativeOS / NativeLanMan deliberately left empty (the smbclient-ng default state).

	if err := c.SessionSetup(&credentials.Credentials{Username: "alice"}); err != nil {
		t.Fatalf("SessionSetup: %v", err)
	}

	reqMsg := message.NewMessage()
	if err := reqMsg.Unmarshal(tr.sent); err != nil {
		t.Fatalf("unmarshal sent request: %v", err)
	}
	req, ok := reqMsg.Command.(*commands.SessionSetupAndxRequest)
	if !ok {
		t.Fatalf("sent command is %T, want *SessionSetupAndxRequest", reqMsg.Command)
	}
	if req.NativeOS != client.DefaultNativeOS {
		t.Errorf("NativeOS = %q, want default %q", req.NativeOS, client.DefaultNativeOS)
	}
	if req.NativeLanMan != client.DefaultNativeLanMan {
		t.Errorf("NativeLanMan = %q, want default %q", req.NativeLanMan, client.DefaultNativeLanMan)
	}
}

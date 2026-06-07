package commands_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
)

// TestFindNotifyCloseDispatch verifies that SMB_COM_FIND_NOTIFY_CLOSE (0x35) is wired
// into the command-casting registry for both requests and responses, so the code is
// represented explicitly instead of reaching the generic "command code not supported"
// default branch.
func TestFindNotifyCloseDispatch(t *testing.T) {
	req, err := commands.CreateRequestCommand(codes.SMB_COM_FIND_NOTIFY_CLOSE)
	if err != nil {
		t.Fatalf("CreateRequestCommand(SMB_COM_FIND_NOTIFY_CLOSE) returned error: %v", err)
	}
	if req == nil {
		t.Fatal("CreateRequestCommand(SMB_COM_FIND_NOTIFY_CLOSE) returned nil command")
	}
	if req.GetCommandCode() != codes.SMB_COM_FIND_NOTIFY_CLOSE {
		t.Errorf("request command code: got %v, want %v", req.GetCommandCode(), codes.SMB_COM_FIND_NOTIFY_CLOSE)
	}

	resp, err := commands.CreateResponseCommand(codes.SMB_COM_FIND_NOTIFY_CLOSE)
	if err != nil {
		t.Fatalf("CreateResponseCommand(SMB_COM_FIND_NOTIFY_CLOSE) returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("CreateResponseCommand(SMB_COM_FIND_NOTIFY_CLOSE) returned nil command")
	}
	if resp.GetCommandCode() != codes.SMB_COM_FIND_NOTIFY_CLOSE {
		t.Errorf("response command code: got %v, want %v", resp.GetCommandCode(), codes.SMB_COM_FIND_NOTIFY_CLOSE)
	}
}

// TestFindNotifyCloseEmptyWireFormat verifies that the never-defined
// SMB_COM_FIND_NOTIFY_CLOSE command carries no parameters and no data: Marshal emits
// only the empty WordCount and ByteCount fields (WordCount=0x00, ByteCount=0x00 0x00),
// and the bytes round-trip through Unmarshal.
func TestFindNotifyCloseEmptyWireFormat(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  interface {
			Marshal() ([]byte, error)
			Unmarshal([]byte) (int, error)
		}
	}{
		{"request", commands.NewFindNotifyCloseRequest()},
		{"response", commands.NewFindNotifyCloseResponse()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			marshalled, err := tc.cmd.Marshal()
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}
			// WordCount (1 byte) = 0x00, then ByteCount (2 bytes) = 0x0000.
			want := []byte{0x00, 0x00, 0x00}
			if len(marshalled) != len(want) {
				t.Fatalf("marshalled length: got %d (% x), want %d (% x)", len(marshalled), marshalled, len(want), want)
			}
			for i := range want {
				if marshalled[i] != want[i] {
					t.Fatalf("marshalled bytes: got % x, want % x", marshalled, want)
				}
			}

			if _, err := tc.cmd.Unmarshal(marshalled); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
		})
	}
}

package client_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestRequestsMatchHiddenAndSystemFiles checks SMB_COM_DELETE, SMB_COM_RENAME and
// SMB_COM_NT_RENAME carry SearchAttributes with the hidden and system bits (and
// the directory bit for the renames): with them clear, the server matches only
// normal files and reports a hidden file as STATUS_NO_SUCH_FILE.
func TestRequestsMatchHiddenAndSystemFiles(t *testing.T) {
	for _, tc := range []struct {
		name string
		resp command_interface.CommandInterface
		call func(*testing.T, *capturingTransport) error
		want uint16
		get  func(command_interface.CommandInterface) (types.SMB_FILE_ATTRIBUTES, bool)
	}{
		{
			name: "DeleteFile", resp: commands.NewDeleteResponse(), want: 0x0006,
			call: func(t *testing.T, tr *capturingTransport) error { return newSessionClient(tr).DeleteFile(`\h.txt`) },
			get: func(c command_interface.CommandInterface) (types.SMB_FILE_ATTRIBUTES, bool) {
				r, ok := c.(*commands.DeleteRequest)
				if !ok {
					return types.SMB_FILE_ATTRIBUTES{}, false
				}
				return r.SearchAttributes, true
			},
		},
		{
			name: "RenameFile", resp: commands.NewRenameResponse(), want: 0x0016,
			call: func(t *testing.T, tr *capturingTransport) error {
				return newSessionClient(tr).RenameFile(`\h.txt`, `\g.txt`)
			},
			get: func(c command_interface.CommandInterface) (types.SMB_FILE_ATTRIBUTES, bool) {
				r, ok := c.(*commands.RenameRequest)
				if !ok {
					return types.SMB_FILE_ATTRIBUTES{}, false
				}
				return r.SearchAttributes, true
			},
		},
		{
			name: "NtRename", resp: commands.NewNtRenameResponse(), want: 0x0016,
			call: func(t *testing.T, tr *capturingTransport) error {
				return newSessionClient(tr).NtRename(`\h.txt`, `\g.txt`)
			},
			get: func(c command_interface.CommandInterface) (types.SMB_FILE_ATTRIBUTES, bool) {
				r, ok := c.(*commands.NtRenameRequest)
				if !ok {
					return types.SMB_FILE_ATTRIBUTES{}, false
				}
				return r.SearchAttributes, true
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tr := &capturingTransport{response: marshalResponse(t, tc.resp)}
			if err := tc.call(t, tr); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			reqMsg := message.NewMessage()
			if err := reqMsg.Unmarshal(tr.sent); err != nil {
				t.Fatalf("unmarshal captured request: %v", err)
			}
			attrs, ok := tc.get(reqMsg.Command)
			if !ok {
				t.Fatalf("captured %T", reqMsg.Command)
			}
			if attrs.Attributes != tc.want {
				t.Errorf("SearchAttributes = 0x%04x, want 0x%04x", attrs.Attributes, tc.want)
			}
		})
	}
}

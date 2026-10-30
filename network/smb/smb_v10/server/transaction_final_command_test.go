package server

import (
	"encoding/binary"
	"testing"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// receiveFinalTransactionResponse reads frames until one that is not an interim
// transaction response, and returns it raw.
//
// A primary that does not carry the whole transaction may be answered with an
// interim response — a success with WordCount and ByteCount of zero — ahead of
// the final one, and only the final one is under test here.
func receiveFinalTransactionResponse(t *testing.T, client *smb1client.Client) []byte {
	t.Helper()

	for {
		raw, err := client.Transport.Receive()
		if err != nil {
			t.Fatalf("the completed transaction was never answered: %v", err)
		}
		interim := len(raw) == 35 && binary.LittleEndian.Uint32(raw[5:9]) == 0 && raw[32] == 0
		if !interim {
			return raw
		}
	}
}

// TestFinalTransactionResponseCarriesThePrimarysCommand asserts the answer to a
// transaction completed by a secondary is coded as the primary's command, not as
// the secondary's.
//
// [MS-CIFS] defines no response to SMB_COM_TRANSACTION2_SECONDARY (section
// 2.2.4.47.2) or SMB_COM_NT_TRANSACT_SECONDARY (section 2.2.4.63.2): what the
// client receives is the response to its primary, and it correlates that answer
// with the command it sent. Samba's server rewrites the command for the same
// reason (Samba bug 8989).
func TestFinalTransactionResponseCarriesThePrimarysCommand(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("file.txt", []byte("contents")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}

	t.Run("TRANSACTION2", func(t *testing.T) {
		_, client := fileServer(t, fs, false)

		// TRANS2_QUERY_FS_INFORMATION's one parameter, the level, split one byte
		// per message.
		primary := commands.NewTransaction2Request()
		primary.Setup = []types.USHORT{types.USHORT(subcommands.TRANS2_QUERY_FS_INFORMATION)}
		primary.SetupCount = 1
		primary.MaxParameterCount = 64
		primary.MaxDataCount = 1024
		primary.TotalParameterCount = 2
		primary.ParameterCount = 1
		primary.Trans2_Parameters = []types.UCHAR{0x03} // SMB_QUERY_FS_SIZE_INFO, low byte
		sendRaw(t, client, primary)

		secondary := commands.NewTransaction2SecondaryRequest()
		secondary.TotalParameterCount = 2
		secondary.ParameterCount = 1
		secondary.ParameterDisplacement = 1
		secondary.Trans2_Parameters = []types.UCHAR{0x01}
		sendRaw(t, client, secondary)

		raw := receiveFinalTransactionResponse(t, client)
		if status := binary.LittleEndian.Uint32(raw[5:9]); status != 0 {
			t.Fatalf("the completed transaction answered 0x%08X, want success", status)
		}
		if got := codes.CommandCode(raw[4]); got != codes.SMB_COM_TRANSACTION2 {
			t.Errorf("the final response is coded 0x%02X, want SMB_COM_TRANSACTION2 (0x32)", uint8(got))
		}
	})

	t.Run("NT_TRANSACT", func(t *testing.T) {
		_, client := fileServer(t, fs, false)
		fid, err := client.OpenFile("file.txt", fileflags.GENERIC_READ|fileflags.READ_CONTROL,
			fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
		if err != nil {
			t.Fatalf("OpenFile() error = %v", err)
		}

		// NT_TRANSACT_QUERY_SECURITY_DESC's parameters — FID(2) Reserved(2)
		// SecurityInfoFields(4) — are split four bytes per message, and the
		// secondary also carries the one data byte the primary declares.
		parameters := make([]byte, 8)
		binary.LittleEndian.PutUint16(parameters[0:2], uint16(fid))
		binary.LittleEndian.PutUint32(parameters[4:8], 0x00000001) // OWNER_SECURITY_INFORMATION

		primary := commands.NewNtTransactRequest()
		primary.Function = types.USHORT(subcommands.NT_TRANSACT_QUERY_SECURITY_DESC)
		primary.MaxParameterCount = 64
		primary.MaxDataCount = 4096
		primary.TotalParameterCount = 8
		primary.ParameterCount = 4
		primary.NT_Trans_Parameters = toUCHARs(parameters[:4])
		primary.TotalDataCount = 1
		sendRaw(t, client, primary)

		secondary := commands.NewNtTransactSecondaryRequest()
		secondary.TotalParameterCount = 8
		secondary.ParameterCount = 4
		secondary.ParameterDisplacement = 4
		secondary.NT_Trans_Parameters = toUCHARs(parameters[4:])
		secondary.TotalDataCount = 1
		secondary.DataCount = 1
		secondary.NT_Trans_Data = []types.UCHAR{0}
		sendRaw(t, client, secondary)

		// The status depends on what the share's security provider answers; the
		// command code does not.
		raw := receiveFinalTransactionResponse(t, client)
		if got := codes.CommandCode(raw[4]); got != codes.SMB_COM_NT_TRANSACT {
			t.Errorf("the final response is coded 0x%02X, want SMB_COM_NT_TRANSACT (0xA0)", uint8(got))
		}
	})
}

// toUCHARs converts a byte slice for a request field.
func toUCHARs(b []byte) []types.UCHAR {
	out := make([]types.UCHAR, len(b))
	for i, v := range b {
		out[i] = types.UCHAR(v)
	}
	return out
}

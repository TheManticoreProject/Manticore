package server

import (
	"encoding/binary"
	"testing"
	"time"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// limitsReplyTimeout bounds the wait for a reply in these tests, so a server that
// never answers fails the test quickly rather than holding the suite.
const limitsReplyTimeout = 2 * time.Second

// receiveLimitsReply reads the reply to a request the test just sent and decodes
// it, failing the test if none arrives: a client that is never answered waits
// forever, which is the defect these tests guard against.
func receiveLimitsReply(t *testing.T, client *smb1client.Client) (*message.Message, uint32) {
	t.Helper()

	client.Transport.SetTimeout(limitsReplyTimeout)
	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("the server never answered the transaction: Receive() error = %v", err)
	}
	if len(raw) < 9 {
		t.Fatalf("the reply is %d bytes", len(raw))
	}
	status := binary.LittleEndian.Uint32(raw[5:9])
	if status != 0 {
		return nil, status
	}
	response := message.NewMessage()
	if err := response.Unmarshal(raw); err != nil {
		t.Fatalf("failed to decode the response: %v", err)
	}
	return response, status
}

// sendTrans2WithLimits sends one TRANSACTION2 with both of the client's receive
// limits under the test's control, and returns the first response and its status.
func sendTrans2WithLimits(
	t *testing.T,
	client *smb1client.Client,
	subcommand subcommands.Transaction2Subcommand,
	parameters []byte,
	maxParameterCount, maxDataCount types.USHORT,
) (*message.Message, uint32) {
	t.Helper()

	request := newRequest(codes.SMB_COM_TRANSACTION2)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID
	request.Header.Flags2 &^= flags2.Flags2(flags2.FLAGS2_UNICODE)

	transaction := commands.NewTransaction2Request()
	transaction.Setup = []types.USHORT{types.USHORT(subcommand)}
	transaction.SetupCount = types.UCHAR(1)
	transaction.MaxParameterCount = maxParameterCount
	transaction.MaxDataCount = maxDataCount
	transaction.TotalParameterCount = types.USHORT(len(parameters))
	transaction.ParameterCount = types.USHORT(len(parameters))
	transaction.Trans2_Parameters = parameters
	request.AddCommand(transaction)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the transaction: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	return receiveLimitsReply(t, client)
}

// sendNtTransactWithLimits sends one NT_TRANSACT with both of the client's
// receive limits under the test's control, and returns the first response and its
// status.
func sendNtTransactWithLimits(
	t *testing.T,
	client *smb1client.Client,
	function subcommands.NtTransactSubcommand,
	parameters []byte,
	maxParameterCount, maxDataCount types.ULONG,
) (*message.Message, uint32) {
	t.Helper()

	request := newRequest(codes.SMB_COM_NT_TRANSACT)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID
	request.Header.Flags2 &^= flags2.Flags2(flags2.FLAGS2_UNICODE)

	transaction := commands.NewNtTransactRequest()
	transaction.Function = types.USHORT(function)
	transaction.MaxParameterCount = maxParameterCount
	transaction.MaxDataCount = maxDataCount
	transaction.TotalParameterCount = types.ULONG(len(parameters))
	transaction.ParameterCount = types.ULONG(len(parameters))
	transaction.NT_Trans_Parameters = parameters
	request.AddCommand(transaction)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the transaction: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	return receiveLimitsReply(t, client)
}

// queryBasicInfoParameters builds a TRANS2_QUERY_PATH_INFORMATION parameter block
// asking for SMB_QUERY_FILE_BASIC_INFO (0x0101): InformationLevel(2)
// Reserved(4) FileName.
func queryBasicInfoParameters(name string) []byte {
	parameters := make([]byte, 6)
	binary.LittleEndian.PutUint16(parameters[0:2], 0x0101)
	return append(parameters, append([]byte(name), 0x00)...)
}

// ntCreateOpenParameters builds an NT_TRANSACT_CREATE parameter block that opens
// an existing file for reading, as TestNtTransactCreateOpensThroughATransaction
// does.
func ntCreateOpenParameters(name string) []byte {
	parameters := make([]byte, 53)
	binary.LittleEndian.PutUint32(parameters[8:12], 0x00120089)  // DesiredAccess: read
	binary.LittleEndian.PutUint32(parameters[28:32], 0x00000001) // FILE_OPEN
	binary.LittleEndian.PutUint32(parameters[32:36], 0x00000040) // FILE_NON_DIRECTORY_FILE
	binary.LittleEndian.PutUint32(parameters[44:48], uint32(len(name)))
	return append(parameters, []byte(name)...)
}

// TestTrans2WithSmallMaxDataCountIsAnswered asserts a TRANSACTION2 whose
// MaxDataCount is smaller than its response parameters is still answered. The
// parameters are bounded by MaxParameterCount, not by MaxDataCount ([MS-CIFS]
// section 2.2.4.46.1), so the client asked for nothing it cannot be given.
func TestTrans2WithSmallMaxDataCountIsAnswered(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("existing.txt", []byte("0123456789")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	response, status := sendTrans2WithLimits(t, client, subcommands.TRANS2_QUERY_PATH_INFORMATION,
		queryBasicInfoParameters("existing.txt"), 1024, 1)
	if status != 0 {
		t.Fatalf("TRANS2_QUERY_PATH_INFORMATION reported 0x%08X, want success", status)
	}
	reply, ok := response.Command.(*commands.Transaction2Response)
	if !ok {
		t.Fatalf("response command is %T", response.Command)
	}
	if len(reply.Trans2_Parameters) != 2 {
		t.Fatalf("the reply carries %d parameter bytes, want the 2-byte EaErrorOffset", len(reply.Trans2_Parameters))
	}
	if len(reply.Trans2_Data) > 1 {
		t.Fatalf("the reply carries %d data bytes, over the client's MaxDataCount of 1", len(reply.Trans2_Data))
	}
}

// TestNtTransactWithSmallMaxDataCountIsAnswered asserts an NT_TRANSACT whose
// MaxDataCount is smaller than its response parameters is still answered: the
// parameters are bounded by MaxParameterCount ([MS-CIFS] section 2.2.4.62.1).
func TestNtTransactWithSmallMaxDataCountIsAnswered(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("existing.txt", []byte("0123456789")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	response, status := sendNtTransactWithLimits(t, client, subcommands.NT_TRANSACT_CREATE,
		ntCreateOpenParameters("existing.txt"), 1024, 10)
	if status != 0 {
		t.Fatalf("NT_TRANSACT_CREATE reported 0x%08X, want success", status)
	}
	reply, ok := response.Command.(*commands.NtTransactResponse)
	if !ok {
		t.Fatalf("response command is %T", response.Command)
	}
	if len(reply.Parameters) != 69 {
		t.Fatalf("the reply carries %d parameter bytes, want the 69 of [MS-CIFS] 2.2.7.1.2", len(reply.Parameters))
	}
}

// TestTrans2AndNtTransactOverMaxParameterCountAreRefused asserts a response whose
// parameters exceed the client's MaxParameterCount is answered with an error
// rather than sent anyway or left unanswered.
func TestTrans2AndNtTransactOverMaxParameterCountAreRefused(t *testing.T) {
	t.Run("TRANS2_QUERY_PATH_INFORMATION", func(t *testing.T) {
		fs := NewMemoryFileSystem("FILES")
		if err := fs.AddFile("existing.txt", []byte("0123456789")); err != nil {
			t.Fatalf("AddFile() error = %v", err)
		}
		_, client := fileServer(t, fs, false)

		_, status := sendTrans2WithLimits(t, client, subcommands.TRANS2_QUERY_PATH_INFORMATION,
			queryBasicInfoParameters("existing.txt"), 1, 4096)
		if status != uint32(nt_status.NT_STATUS_INVALID_PARAMETER) {
			t.Fatalf("the query reported 0x%08X, want STATUS_INVALID_PARAMETER", status)
		}
	})

	t.Run("NT_TRANSACT_CREATE", func(t *testing.T) {
		fs := NewMemoryFileSystem("FILES")
		if err := fs.AddFile("existing.txt", []byte("0123456789")); err != nil {
			t.Fatalf("AddFile() error = %v", err)
		}
		_, client := fileServer(t, fs, false)

		_, status := sendNtTransactWithLimits(t, client, subcommands.NT_TRANSACT_CREATE,
			ntCreateOpenParameters("existing.txt"), 68, 4096)
		if status != uint32(nt_status.NT_STATUS_INVALID_PARAMETER) {
			t.Fatalf("the create reported 0x%08X, want STATUS_INVALID_PARAMETER", status)
		}
	})
}

package server

import (
	"testing"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// sendPipeTransactionWithLimits sends one SMB_COM_TRANSACTION against a pipe
// handle with both of the client's receive limits under the test's control, and
// returns the decoded response and its status.
//
// It fails the test if no response arrives at all, which is the defect these
// tests guard against: a client that is never answered waits forever.
func sendPipeTransactionWithLimits(
	t *testing.T,
	client *smb1client.Client,
	subcommand subcommands.TransactionSubcommand,
	fid smb1client.FID,
	maxParameterCount, maxDataCount types.USHORT,
) (*message.Message, uint32) {
	t.Helper()

	request := newRequest(codes.SMB_COM_TRANSACTION)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID
	request.Header.Flags2 &^= flags2.Flags2(flags2.FLAGS2_UNICODE)

	transaction := commands.NewTransactionRequest()
	transaction.Setup = []types.USHORT{types.USHORT(subcommand), types.USHORT(fid)}
	transaction.SetupCount = types.UCHAR(len(transaction.Setup))
	transaction.MaxParameterCount = maxParameterCount
	transaction.MaxDataCount = maxDataCount
	if err := transaction.Name.SetString(`\PIPE\`); err != nil {
		t.Fatalf("Name.SetString() error = %v", err)
	}
	request.AddCommand(transaction)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the transaction: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("the server never answered the transaction: Receive() error = %v", err)
	}
	response := message.NewMessage()
	if err := response.Unmarshal(raw); err != nil {
		t.Fatalf("failed to decode the response: %v", err)
	}
	return response, response.Header.Status
}

// TestPipePeekWithSmallMaxDataCountIsAnswered asserts a TRANS_PEEK_NMPIPE whose
// MaxDataCount is smaller than the 6 parameter bytes every peek response carries
// is still answered. The parameters are bounded by MaxParameterCount, not by
// MaxDataCount, so the client asked for nothing it cannot be given.
func TestPipePeekWithSmallMaxDataCountIsAnswered(t *testing.T) {
	_, client := pipeServer(t, &oversizedAnswerPipe{name: "srvsvc", answer: countingAnswer(200)})

	fid, err := openPipeHandle(t, client, `\PIPE\srvsvc`)
	if err != nil {
		t.Fatalf("opening the pipe failed: %v", err)
	}

	response, status := sendPipeTransactionWithLimits(t, client, subcommands.TRANS_PEEK_NMPIPE, fid, 1024, 3)
	if status != 0 {
		t.Fatalf("the peek reported 0x%08X, want success", status)
	}
	peek, ok := response.Command.(*commands.TransactionResponse)
	if !ok {
		t.Fatalf("response command is %T", response.Command)
	}
	if len(peek.Trans_Parameters) != peekParametersSize {
		t.Fatalf("the peek returned %d parameter bytes, want %d", len(peek.Trans_Parameters), peekParametersSize)
	}
	if len(peek.Trans_Data) > 3 {
		t.Fatalf("the peek returned %d data bytes, over the client's MaxDataCount of 3", len(peek.Trans_Data))
	}
}

// TestPipeResponseOverMaxParameterCountIsRefused asserts a pipe operation whose
// response parameters exceed the client's MaxParameterCount is answered with an
// error rather than left unanswered: the server must not send more parameter
// bytes than the client accepts, and silence would leave the client waiting.
func TestPipeResponseOverMaxParameterCountIsRefused(t *testing.T) {
	cases := []struct {
		name              string
		subcommand        subcommands.TransactionSubcommand
		maxParameterCount types.USHORT
	}{
		{"TRANS_PEEK_NMPIPE", subcommands.TRANS_PEEK_NMPIPE, peekParametersSize - 1},
		{"TRANS_QUERY_NMPIPE_STATE", subcommands.TRANS_QUERY_NMPIPE_STATE, 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, client := pipeServer(t, &oversizedAnswerPipe{name: "srvsvc", answer: countingAnswer(200)})

			fid, err := openPipeHandle(t, client, `\PIPE\srvsvc`)
			if err != nil {
				t.Fatalf("opening the pipe failed: %v", err)
			}

			_, status := sendPipeTransactionWithLimits(t, client, tc.subcommand, fid, tc.maxParameterCount, 4096)
			if status != uint32(nt_status.NT_STATUS_INVALID_PARAMETER) {
				t.Fatalf("the operation reported 0x%08X, want STATUS_INVALID_PARAMETER", status)
			}
		})
	}
}

package server

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// TestTrans2DataExactlyMaxDataCountIsOneResponse asserts a TRANSACTION2 reply
// whose data is exactly the client's MaxDataCount arrives whole in one response.
// MaxDataCount bounds the data alone ([MS-CIFS] section 2.2.4.46.1), so the
// 2-byte EaErrorOffset in the parameters must not push the tail of the data into
// a second message the client did not ask for.
func TestTrans2DataExactlyMaxDataCountIsOneResponse(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("existing.txt", []byte("0123456789")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	// The first query learns the size of the data; the second asks for exactly
	// that much.
	sized, status := sendTrans2WithLimits(t, client, subcommands.TRANS2_QUERY_PATH_INFORMATION,
		queryBasicInfoParameters("existing.txt"), 1024, 4096)
	if status != 0 {
		t.Fatalf("TRANS2_QUERY_PATH_INFORMATION reported 0x%08X, want success", status)
	}
	sizedReply, ok := sized.Command.(*commands.Transaction2Response)
	if !ok {
		t.Fatalf("response command is %T", sized.Command)
	}
	dataLength := len(sizedReply.Trans2_Data)
	if dataLength == 0 || len(sizedReply.Trans2_Parameters) == 0 {
		t.Fatalf("the reply carries %d parameter and %d data bytes, want both", len(sizedReply.Trans2_Parameters), dataLength)
	}

	response, status := sendTrans2WithLimits(t, client, subcommands.TRANS2_QUERY_PATH_INFORMATION,
		queryBasicInfoParameters("existing.txt"), 1024, types.USHORT(dataLength))
	if status != 0 {
		t.Fatalf("TRANS2_QUERY_PATH_INFORMATION reported 0x%08X, want success", status)
	}
	reply, ok := response.Command.(*commands.Transaction2Response)
	if !ok {
		t.Fatalf("response command is %T", response.Command)
	}
	if int(reply.TotalDataCount) != dataLength {
		t.Fatalf("TotalDataCount is %d, want %d", reply.TotalDataCount, dataLength)
	}
	if len(reply.Trans2_Data) != dataLength {
		t.Fatalf("the first response carries %d of %d data bytes, want all of them with a MaxDataCount of %d",
			len(reply.Trans2_Data), dataLength, dataLength)
	}
	if !bytes.Equal(reply.Trans2_Data, sizedReply.Trans2_Data) {
		t.Fatal("the data differs from the unconstrained reply")
	}
}

// TestPipePeekReturnsMaxDataCountBytes asserts a TRANS_PEEK_NMPIPE with a
// MaxDataCount of N returns N bytes of the buffered answer in one response. The
// 6 parameter bytes are bounded by MaxParameterCount, not by MaxDataCount
// ([MS-CIFS] section 2.2.4.33.1), so they do not come out of the data budget.
func TestPipePeekReturnsMaxDataCountBytes(t *testing.T) {
	const transactBudget = 64
	const peekBudget = 32
	answer := countingAnswer(305)

	_, client := pipeServer(t, &oversizedAnswerPipe{name: "srvsvc", answer: answer})

	fid, err := openPipeHandle(t, client, `\PIPE\srvsvc`)
	if err != nil {
		t.Fatalf("opening the pipe failed: %v", err)
	}

	// A transaction with a small buffer leaves the rest of the answer on the
	// handle for the peek to look at.
	setup := []types.USHORT{types.USHORT(subcommands.TRANS_TRANSACT_NMPIPE), types.USHORT(fid)}
	_, status := sendTransactionWithLimit(t, client, setup, `\PIPE\`, nil, []byte("bind"), transactBudget)
	if status != uint32(nt_status.NT_STATUS_BUFFER_OVERFLOW) {
		t.Fatalf("the transaction reported 0x%08X, want STATUS_BUFFER_OVERFLOW", status)
	}
	buffered := len(answer) - transactBudget

	response, status := sendPipeTransactionWithLimits(t, client, subcommands.TRANS_PEEK_NMPIPE, fid, 1024, peekBudget)
	if status != uint32(nt_status.NT_STATUS_BUFFER_OVERFLOW) {
		t.Fatalf("the peek reported 0x%08X, want STATUS_BUFFER_OVERFLOW", status)
	}
	peek, ok := response.Command.(*commands.TransactionResponse)
	if !ok {
		t.Fatalf("response command is %T", response.Command)
	}
	if len(peek.Trans_Parameters) != peekParametersSize {
		t.Fatalf("the peek returned %d parameter bytes, want %d", len(peek.Trans_Parameters), peekParametersSize)
	}
	if len(peek.Trans_Data) != peekBudget || int(peek.TotalDataCount) != peekBudget {
		t.Fatalf("the peek returned %d data bytes of a total of %d, want %d with a MaxDataCount of %d",
			len(peek.Trans_Data), peek.TotalDataCount, peekBudget, peekBudget)
	}
	if !bytes.Equal(peek.Trans_Data, answer[transactBudget:transactBudget+peekBudget]) {
		t.Fatal("the peeked data is not the front of the buffered answer")
	}

	parameters := []byte(peek.Trans_Parameters)
	if available := binary.LittleEndian.Uint16(parameters[0:2]); int(available) != buffered {
		t.Errorf("ReadDataAvailable is %d, want %d", available, buffered)
	}
	if remaining := binary.LittleEndian.Uint16(parameters[2:4]); int(remaining) != buffered-peekBudget {
		t.Errorf("MessageBytesLength is %d, want %d", remaining, buffered-peekBudget)
	}
}

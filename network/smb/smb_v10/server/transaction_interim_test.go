package server

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// expectInterimResponse reads the next frame and asserts it is the interim
// response to a primary transaction request: the primary's command code, a
// success status, and WordCount and ByteCount of zero.
//
// It reads the raw frame rather than decoding it, because a transaction response
// with no words does not decode as the full response structure.
func expectInterimResponse(t *testing.T, transport interface{ Receive() ([]byte, error) }, command codes.CommandCode) {
	t.Helper()

	raw, err := transport.Receive()
	if err != nil {
		t.Fatalf("no interim response arrived: %v", err)
	}
	// Header(32) WordCount(1) ByteCount(2).
	if len(raw) != 35 {
		t.Fatalf("the interim response is %d bytes (% x), want 35", len(raw), raw)
	}
	if got := codes.CommandCode(raw[4]); got != command {
		t.Errorf("the interim response carries command 0x%02X, want 0x%02X", uint8(got), uint8(command))
	}
	if status := binary.LittleEndian.Uint32(raw[5:9]); status != 0 {
		t.Errorf("the interim response carries status 0x%08X, want success", status)
	}
	if raw[32] != 0 || binary.LittleEndian.Uint16(raw[33:35]) != 0 {
		t.Errorf("the interim response has WordCount %d and ByteCount %d, want 0 and 0",
			raw[32], binary.LittleEndian.Uint16(raw[33:35]))
	}
}

// TestFragmentedTransaction2IsAnsweredWithAnInterimResponse asserts a primary
// TRANSACTION2 that does not carry the whole transaction is answered with an
// interim response, and that the transaction then completes once the secondary
// brings the rest.
//
// [MS-CIFS] section 3.3.5.2.5: when the primary's counts fall short of its totals
// "the server MUST send an Interim Response message, setting the
// SMB_Parameters.WordCount and SMB_Data.ByteCount fields to 0". A client sends no
// secondary until that arrives, so without it the transaction never completes.
func TestFragmentedTransaction2IsAnsweredWithAnInterimResponse(t *testing.T) {
	_, client := fileServer(t, NewMemoryFileSystem("FILES"), false)

	// TRANS2_QUERY_FS_INFORMATION carries one parameter, the two-byte level,
	// which is split one byte per message.
	level := []byte{0x03, 0x01} // SMB_QUERY_FS_SIZE_INFO

	transaction := commands.NewTransaction2Request()
	transaction.Setup = []types.USHORT{types.USHORT(subcommands.TRANS2_QUERY_FS_INFORMATION)}
	transaction.SetupCount = 1
	transaction.MaxParameterCount = 64
	transaction.MaxDataCount = 1024
	transaction.TotalParameterCount = 2
	transaction.ParameterCount = 1
	transaction.Trans2_Parameters = []types.UCHAR{types.UCHAR(level[0])}
	sendRaw(t, client, transaction)

	expectInterimResponse(t, client.Transport, codes.SMB_COM_TRANSACTION2)

	continuation := commands.NewTransaction2SecondaryRequest()
	continuation.TotalParameterCount = 2
	continuation.ParameterCount = 1
	continuation.ParameterDisplacement = 1
	continuation.Trans2_Parameters = []types.UCHAR{types.UCHAR(level[1])}
	sendRaw(t, client, continuation)

	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("the completed transaction was never answered: %v", err)
	}
	if status := binary.LittleEndian.Uint32(raw[5:9]); status != 0 {
		t.Fatalf("the completed transaction answered 0x%08X, want success", status)
	}
	if raw[32] == 0 {
		t.Fatal("the completed transaction was answered with another interim response")
	}
}

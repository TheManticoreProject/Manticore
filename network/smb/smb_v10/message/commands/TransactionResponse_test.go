package commands_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestTransactionResponseRoundTrip verifies that TransactionResponse serializes its
// parameter and data payloads and recovers them on Unmarshal, with the transaction
// counts preserved.
func TestTransactionResponseRoundTrip(t *testing.T) {
	transParams := []byte{0x10, 0x20, 0x30, 0x40}
	transData := []byte("transaction payload bytes")

	resp := commands.NewTransactionResponse()
	resp.TotalParameterCount = types.USHORT(len(transParams))
	resp.ParameterCount = types.USHORT(len(transParams))
	resp.TotalDataCount = types.USHORT(len(transData))
	resp.DataCount = types.USHORT(len(transData))
	resp.Setup = []types.USHORT{0x00AA, 0x00BB}
	resp.SetupCount = types.UCHAR(len(resp.Setup))
	resp.Trans_Parameters = []types.UCHAR(transParams)
	resp.Trans_Data = []types.UCHAR(transData)

	raw, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	decoded := commands.NewTransactionResponse()
	if _, err := decoded.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.TotalParameterCount != types.USHORT(len(transParams)) {
		t.Errorf("TotalParameterCount = %d, want %d", decoded.TotalParameterCount, len(transParams))
	}
	if decoded.TotalDataCount != types.USHORT(len(transData)) {
		t.Errorf("TotalDataCount = %d, want %d", decoded.TotalDataCount, len(transData))
	}
	if int(decoded.SetupCount) != 2 || len(decoded.Setup) != 2 || decoded.Setup[0] != 0x00AA || decoded.Setup[1] != 0x00BB {
		t.Errorf("Setup = %v (count %d), want [0x00AA 0x00BB]", decoded.Setup, decoded.SetupCount)
	}
	if !bytes.Equal([]byte(decoded.Trans_Parameters), transParams) {
		t.Errorf("Trans_Parameters = % x, want % x", []byte(decoded.Trans_Parameters), transParams)
	}
	if !bytes.Equal([]byte(decoded.Trans_Data), transData) {
		t.Errorf("Trans_Data = % x, want % x", []byte(decoded.Trans_Data), transData)
	}
}

// TestTransactionResponseInterimIsEmpty verifies that a freshly constructed response
// (representing the interim response with empty Parameter and Data sections) round-trips
// without error and yields empty payloads.
func TestTransactionResponseInterimIsEmpty(t *testing.T) {
	resp := commands.NewTransactionResponse()
	raw, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	decoded := commands.NewTransactionResponse()
	if _, err := decoded.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(decoded.Trans_Parameters) != 0 || len(decoded.Trans_Data) != 0 {
		t.Errorf("expected empty payloads, got params=%d data=%d", len(decoded.Trans_Parameters), len(decoded.Trans_Data))
	}
}

// TestTransactionResponseUnmarshalHeaderRelativeOffsets verifies that
// TransactionResponse.Unmarshal interprets ParameterOffset and DataOffset as offsets
// from the start of the SMB Header (per [MS-CIFS] 2.2.4.33.2) when deriving the
// Pad1/Pad2 alignment lengths, matching what a real server emits.
func TestTransactionResponseUnmarshalHeaderRelativeOffsets(t *testing.T) {
	const wordCount = 10

	transParams := []byte{0x01, 0x08, 0x05, 0x00}
	transData := []byte{0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7}

	// Layout from the start of the SMB Header:
	//   Header(32) WordCount(1) words(20) ByteCount(2) -> Bytes block starts at 55.
	const bytesBlockStart = 32 + 1 + 2*wordCount + 2 // 55
	pad1 := (4 - (bytesBlockStart % 4)) % 4
	paramOffset := bytesBlockStart + pad1
	afterParams := paramOffset + len(transParams)
	pad2 := (4 - (afterParams % 4)) % 4
	dataOffset := afterParams + pad2

	words := make([]byte, 2*wordCount)
	binary.LittleEndian.PutUint16(words[0:2], uint16(len(transParams)))  // TotalParameterCount
	binary.LittleEndian.PutUint16(words[2:4], uint16(len(transData)))    // TotalDataCount
	binary.LittleEndian.PutUint16(words[6:8], uint16(len(transParams)))  // ParameterCount
	binary.LittleEndian.PutUint16(words[8:10], uint16(paramOffset))      // ParameterOffset (header-relative)
	binary.LittleEndian.PutUint16(words[12:14], uint16(len(transData)))  // DataCount
	binary.LittleEndian.PutUint16(words[14:16], uint16(dataOffset))      // DataOffset (header-relative)

	bytesBlock := []byte{}
	bytesBlock = append(bytesBlock, make([]byte, pad1)...)
	bytesBlock = append(bytesBlock, transParams...)
	bytesBlock = append(bytesBlock, make([]byte, pad2)...)
	bytesBlock = append(bytesBlock, transData...)

	raw := []byte{byte(wordCount)}
	raw = append(raw, words...)
	bc := make([]byte, 2)
	binary.LittleEndian.PutUint16(bc, uint16(len(bytesBlock)))
	raw = append(raw, bc...)
	raw = append(raw, bytesBlock...)

	resp := commands.NewTransactionResponse()
	if _, err := resp.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if !bytes.Equal([]byte(resp.Trans_Parameters), transParams) {
		t.Errorf("Trans_Parameters = % x, want % x", []byte(resp.Trans_Parameters), transParams)
	}
	if !bytes.Equal([]byte(resp.Trans_Data), transData) {
		t.Errorf("Trans_Data = % x, want % x", []byte(resp.Trans_Data), transData)
	}
}

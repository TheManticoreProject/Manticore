package commands_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
)

// TestTransaction2ResponseUnmarshalHeaderRelativeOffsets verifies that
// Transaction2Response.Unmarshal interprets ParameterOffset and DataOffset as
// offsets from the start of the SMB Header (per [MS-CIFS] 2.2.4.46.2) when
// deriving the Pad1/Pad2 alignment lengths. A real server response places the
// data block after the 32-byte header; omitting the header size from that
// computation shifts every payload run and overruns Trans2_Data.
func TestTransaction2ResponseUnmarshalHeaderRelativeOffsets(t *testing.T) {
	const wordCount = 10

	trans2Params := []byte{0x01, 0x08, 0x05, 0x00} // 4 bytes (e.g. FIND_FIRST2 reply params)
	trans2Data := []byte{0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7}

	// Layout from the start of the SMB Header:
	//   Header(32) WordCount(1) words(20) ByteCount(2) -> Bytes block starts at 55.
	const bytesBlockStart = 32 + 1 + 2*wordCount + 2 // 55
	pad1 := (4 - (bytesBlockStart % 4)) % 4          // align ParameterOffset to 4
	paramOffset := bytesBlockStart + pad1
	afterParams := paramOffset + len(trans2Params)
	pad2 := (4 - (afterParams % 4)) % 4
	dataOffset := afterParams + pad2

	// SMB_Parameters words (header-relative offsets are written into the words).
	words := make([]byte, 2*wordCount)
	binary.LittleEndian.PutUint16(words[0:2], uint16(len(trans2Params))) // TotalParameterCount
	binary.LittleEndian.PutUint16(words[2:4], uint16(len(trans2Data)))   // TotalDataCount
	binary.LittleEndian.PutUint16(words[6:8], uint16(len(trans2Params))) // ParameterCount
	binary.LittleEndian.PutUint16(words[8:10], uint16(paramOffset))      // ParameterOffset (header-relative)
	binary.LittleEndian.PutUint16(words[12:14], uint16(len(trans2Data))) // DataCount
	binary.LittleEndian.PutUint16(words[14:16], uint16(dataOffset))      // DataOffset (header-relative)
	// SetupCount (byte 18) and Reserved2 (byte 19) stay zero.

	// SMB_Data.Bytes block: Pad1 | Trans2_Parameters | Pad2 | Trans2_Data.
	bytesBlock := []byte{}
	bytesBlock = append(bytesBlock, make([]byte, pad1)...)
	bytesBlock = append(bytesBlock, trans2Params...)
	bytesBlock = append(bytesBlock, make([]byte, pad2)...)
	bytesBlock = append(bytesBlock, trans2Data...)

	// The command bytes as handed to Unmarshal (everything after the SMB Header):
	// WordCount | words | ByteCount | Bytes block.
	raw := []byte{byte(wordCount)}
	raw = append(raw, words...)
	bc := make([]byte, 2)
	binary.LittleEndian.PutUint16(bc, uint16(len(bytesBlock)))
	raw = append(raw, bc...)
	raw = append(raw, bytesBlock...)

	resp := commands.NewTransaction2Response()
	if _, err := resp.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !bytes.Equal([]byte(resp.Trans2_Parameters), trans2Params) {
		t.Errorf("Trans2_Parameters = % x, want % x", []byte(resp.Trans2_Parameters), trans2Params)
	}
	if !bytes.Equal([]byte(resp.Trans2_Data), trans2Data) {
		t.Errorf("Trans2_Data = % x, want % x", []byte(resp.Trans2_Data), trans2Data)
	}
}

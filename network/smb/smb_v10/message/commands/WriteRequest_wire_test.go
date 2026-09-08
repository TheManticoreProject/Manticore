package commands

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestWriteRequestPutsItsDataInTheDataBlock asserts the command marshals as
// WordCount, the parameter words, ByteCount and then the data — the order
// [MS-CIFS] section 2.2.4.12.1 gives it.
//
// The data field once went into the command buffer rather than the data block,
// which emitted the payload ahead of WordCount and left the data block empty. The
// result was not merely mis-sized: a decoder read the leading buffer-format byte
// as WordCount, so the message could not be parsed at all.
func TestWriteRequestPutsItsDataInTheDataBlock(t *testing.T) {
	payload := []byte("bytes to write")

	request := NewWriteRequest()
	request.FID = types.USHORT(0x0042)
	request.CountOfBytesToWrite = types.USHORT(len(payload))
	request.WriteOffsetInBytes = types.ULONG(0x11223344)
	request.Data.Buffer = []types.UCHAR(payload)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// WordCount is five words: FID(2) CountOfBytesToWrite(2) WriteOffsetInBytes(4)
	// EstimateOfRemainingBytesToBeWritten(2).
	const wantWords = 5
	if got := int(marshalled[0]); got != wantWords {
		t.Fatalf("the command begins with 0x%02X as WordCount, want %d — the data block is in front of it",
			marshalled[0], wantWords)
	}

	if fid := binary.LittleEndian.Uint16(marshalled[1:3]); fid != 0x0042 {
		t.Errorf("FID is 0x%04X at the first parameter word, want 0x0042", fid)
	}
	if offset := binary.LittleEndian.Uint32(marshalled[5:9]); offset != 0x11223344 {
		t.Errorf("WriteOffsetInBytes is 0x%08X, want 0x11223344", offset)
	}

	// ByteCount follows the words and has to describe the data block that
	// follows it: BufferFormat(1) DataLength(2) Data.
	byteCountAt := 1 + 2*wantWords
	byteCount := int(binary.LittleEndian.Uint16(marshalled[byteCountAt : byteCountAt+2]))
	wantByteCount := 1 + 2 + len(payload)
	if byteCount != wantByteCount {
		t.Fatalf("ByteCount is %d, want %d — the data block is empty", byteCount, wantByteCount)
	}
	if len(marshalled) != byteCountAt+2+wantByteCount {
		t.Fatalf("the command is %d bytes, want %d", len(marshalled), byteCountAt+2+wantByteCount)
	}

	block := marshalled[byteCountAt+2:]
	if block[0] != types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT {
		t.Errorf("the data block begins with BufferFormat 0x%02X, want 0x01", block[0])
	}
	if declared := int(binary.LittleEndian.Uint16(block[1:3])); declared != len(payload) {
		t.Errorf("the data block declares %d bytes, want %d", declared, len(payload))
	}
	if got := string(block[3:]); got != string(payload) {
		t.Errorf("the data block carries %q, want %q", got, payload)
	}
}

// TestWriteRequestRoundTrips asserts the command decodes what it encodes, which it
// could not while its payload sat in front of the word count.
func TestWriteRequestRoundTrips(t *testing.T) {
	payload := []byte("bytes to write")

	request := NewWriteRequest()
	request.FID = types.USHORT(0x0042)
	request.CountOfBytesToWrite = types.USHORT(len(payload))
	request.WriteOffsetInBytes = types.ULONG(0x11223344)
	request.Data.Buffer = []types.UCHAR(payload)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	decoded := NewWriteRequest()
	if _, err := decoded.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.FID != request.FID {
		t.Errorf("FID round-tripped to 0x%04X, want 0x%04X", decoded.FID, request.FID)
	}
	if decoded.CountOfBytesToWrite != request.CountOfBytesToWrite {
		t.Errorf("CountOfBytesToWrite round-tripped to %d, want %d",
			decoded.CountOfBytesToWrite, request.CountOfBytesToWrite)
	}
	if decoded.WriteOffsetInBytes != request.WriteOffsetInBytes {
		t.Errorf("WriteOffsetInBytes round-tripped to 0x%08X, want 0x%08X",
			decoded.WriteOffsetInBytes, request.WriteOffsetInBytes)
	}
	if got := string(decoded.Data.Buffer); got != string(payload) {
		t.Errorf("the data round-tripped to %q, want %q", got, payload)
	}
}

package commands

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// countingPayload is a block whose every byte identifies its own position, so a
// transfer that is right in length but wrong in content, or reassembled out of
// order, still fails.
func countingPayload(length int) []byte {
	block := make([]byte, length)
	for index := range block {
		block[index] = byte(index % 251)
	}
	return block
}

// The positions of the fields this file patches, in bytes from the start of a
// WriteAndx request's own block: WordCount(1), then the AndX block(4), FID(2),
// Offset(4), Timeout(4), WriteMode(2), Remaining(2), and then the three fields
// that describe the data.
const (
	writeAndxDataLengthHighAt = 1 + 4 + 2 + 4 + 4 + 2 + 2
	writeAndxDataLengthAt     = writeAndxDataLengthHighAt + 2
	writeAndxDataOffsetAt     = writeAndxDataLengthAt + 2
	writeAndxByteCountAt      = writeAndxDataOffsetAt + 2
)

func TestWriteAndxRequestDerivesTheFieldsThatDescribeItsData(t *testing.T) {
	// A caller sets Data and nothing else: the length and the offset are the
	// command's own wire layout, which the command is the only thing that knows.
	payload := countingPayload(300)

	request := NewWriteAndxRequest()
	request.FID = types.USHORT(0x4001)
	request.Data = []types.UCHAR(payload)

	if _, err := request.Marshal(); err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if int(request.DataLength) != len(payload) {
		t.Errorf("DataLength is %d, want %d", request.DataLength, len(payload))
	}
	if request.DataLengthHigh() != 0 {
		t.Errorf("DataLengthHigh is %d for a %d-byte write, want 0", request.DataLengthHigh(), len(payload))
	}
	if request.DataOffset != writeAndxDataOffset {
		t.Errorf("DataOffset is %d, want %d", request.DataOffset, writeAndxDataOffset)
	}
}

func TestWriteAndxRequestAbove64KiBRoundTrips(t *testing.T) {
	// The case the length fields exist for: a write no single USHORT can describe.
	payload := countingPayload(0x10000 + 1234)

	request := NewWriteAndxRequest()
	request.FID = types.USHORT(0x1234)
	request.Offset = types.ULONG(4096)
	request.Data = []types.UCHAR(payload)

	raw, err := request.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// The length is split across the two words, and neither alone describes it.
	if want := types.USHORT(len(payload) & 0xFFFF); request.DataLength != want {
		t.Errorf("DataLength is 0x%04X, want 0x%04X", request.DataLength, want)
	}
	if want := types.USHORT(len(payload) >> 16); request.DataLengthHigh() != want {
		t.Errorf("DataLengthHigh is 0x%04X, want 0x%04X", request.DataLengthHigh(), want)
	}

	// ByteCount carries the low 16 bits of the block, which is all a USHORT can
	// hold. Asserting it rather than ignoring it records that the value is
	// deliberate: a receiver of a write this size reads the length fields, which
	// is what [MS-SMB] section 3.3.5.8 has Windows servers do.
	block := 1 + len(payload)
	if got := binary.LittleEndian.Uint16(raw[writeAndxByteCountAt : writeAndxByteCountAt+2]); got != uint16(block&0xFFFF) {
		t.Errorf("ByteCount is 0x%04X, want the low word of the %d-byte block (0x%04X)",
			got, block, uint16(block&0xFFFF))
	}

	decoded := NewWriteAndxRequest()
	if _, err := decoded.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(decoded.Data) != len(payload) {
		t.Fatalf("decoded %d bytes, want %d", len(decoded.Data), len(payload))
	}
	if !bytes.Equal([]byte(decoded.Data), payload) {
		t.Error("the decoded bytes are not the ones written")
	}
}

func TestWriteAndxRequestSixtyFourBitFormMovesItsData(t *testing.T) {
	// The 64-bit-offset form carries OffsetHigh in its parameter block, which
	// pushes the data four bytes further from the header. A DataOffset that did
	// not follow would point four bytes short and the server would write the
	// wrong bytes.
	payload := countingPayload(0x10000)

	request := NewWriteAndxRequest()
	request.FID = types.USHORT(7)
	request.Offset = types.ULONG(0x80000000)
	request.OffsetHigh = types.ULONG(1)
	request.Data = []types.UCHAR(payload)

	raw, err := request.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if request.DataOffset != writeAndxDataOffset64 {
		t.Errorf("DataOffset is %d, want the 64-bit form's %d", request.DataOffset, writeAndxDataOffset64)
	}

	decoded := NewWriteAndxRequest()
	if _, err := decoded.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if decoded.OffsetHigh != 1 {
		t.Errorf("decoded OffsetHigh is %d, want 1", decoded.OffsetHigh)
	}
	if !bytes.Equal([]byte(decoded.Data), payload) {
		t.Error("the decoded bytes are not the ones written")
	}
}

func TestWriteAndxRequestBatchedResolvesItsDataOffset(t *testing.T) {
	// Batched behind another command, DataOffset has to account for everything
	// ahead of it: it is measured from the SMB header "regardless of the command
	// request's position in an AndX chain" ([MS-CIFS] section 2.2.4.43.1).
	const chainOffset = 41
	payload := countingPayload(0x10000 + 8)

	request := NewWriteAndxRequest()
	request.FID = types.USHORT(9)
	request.Data = []types.UCHAR(payload)
	request.SetChainOffset(chainOffset)

	raw, err := request.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if want := types.USHORT(writeAndxDataOffset + chainOffset); request.DataOffset != want {
		t.Errorf("DataOffset is %d, want %d", request.DataOffset, want)
	}

	// Decoded without being told where it sits, the offset resolves 41 bytes past
	// the data and the request is refused rather than silently reading the wrong
	// window.
	unpositioned := NewWriteAndxRequest()
	if _, err := unpositioned.Unmarshal(raw); err == nil {
		t.Error("a batched request decoded as if it were first in the message")
	}

	decoded := NewWriteAndxRequest()
	decoded.SetChainOffset(chainOffset)
	if _, err := decoded.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if !bytes.Equal([]byte(decoded.Data), payload) {
		t.Error("the decoded bytes are not the ones written")
	}
}

func TestWriteAndxRequestRejectsADataOffsetItsOwnParametersCover(t *testing.T) {
	// A DataOffset inside the command's own parameter block is not a relocation:
	// [MS-CIFS] section 3.3.5.37 has the server fail such a request, and an error
	// here becomes STATUS_INVALID_SMB.
	request := NewWriteAndxRequest()
	request.FID = types.USHORT(1)
	request.Data = []types.UCHAR(countingPayload(64))

	raw, err := request.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Point it at the FID, which is inside the parameter words.
	binary.LittleEndian.PutUint16(raw[writeAndxDataOffsetAt:writeAndxDataOffsetAt+2],
		uint16(header.SMB_HEADER_SIZE+1+4))

	decoded := NewWriteAndxRequest()
	if _, err := decoded.Unmarshal(raw); err == nil {
		t.Error("a DataOffset pointing into the parameter words was accepted")
	}
}

func TestWriteAndxRequestRejectsDataPastTheEndOfTheMessage(t *testing.T) {
	// A declared length that runs off the end of what arrived is a malformed
	// request, not licence to read past the buffer.
	request := NewWriteAndxRequest()
	request.FID = types.USHORT(1)
	request.Data = []types.UCHAR(countingPayload(64))

	raw, err := request.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Claim a megabyte, across both words, of a message that carries 64 bytes.
	binary.LittleEndian.PutUint16(raw[writeAndxDataLengthAt:writeAndxDataLengthAt+2], 0x0000)
	binary.LittleEndian.PutUint16(raw[writeAndxDataLengthHighAt:writeAndxDataLengthHighAt+2], 0x0010)

	decoded := NewWriteAndxRequest()
	if _, err := decoded.Unmarshal(raw); err == nil {
		t.Error("a request declaring a megabyte of data in a 64-byte message was accepted")
	}
}

func TestWriteAndxRequestFindsRelocatedData(t *testing.T) {
	// The relocation [MS-CIFS] section 2.2.4.43.1 describes: the data sits past
	// the end of the command's own block, with a pad between. Nothing but
	// DataOffset can find it.
	payload := countingPayload(0x10000 + 3)

	request := NewWriteAndxRequest()
	request.FID = types.USHORT(1)

	raw, err := request.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Seven bytes of filler stand in for whatever the sender put between its
	// command block and the relocated data — another command in the chain, or
	// alignment padding.
	relocated := append(append([]byte{}, raw...), make([]byte, 7)...)
	dataAt := len(relocated)
	relocated = append(relocated, payload...)

	binary.LittleEndian.PutUint16(relocated[writeAndxDataLengthAt:writeAndxDataLengthAt+2],
		uint16(len(payload)&0xFFFF))
	binary.LittleEndian.PutUint16(relocated[writeAndxDataLengthHighAt:writeAndxDataLengthHighAt+2],
		uint16(len(payload)>>16))
	binary.LittleEndian.PutUint16(relocated[writeAndxDataOffsetAt:writeAndxDataOffsetAt+2],
		uint16(header.SMB_HEADER_SIZE+dataAt))
	// ByteCount stays at "1 + SMB_Parameters.Words.DataLength", as the
	// specification's own example has it — counting bytes that are not in the
	// block. A decoder that trusted it would refuse this message.
	binary.LittleEndian.PutUint16(relocated[writeAndxByteCountAt:writeAndxByteCountAt+2],
		uint16((1+len(payload))&0xFFFF))

	decoded := NewWriteAndxRequest()
	if _, err := decoded.Unmarshal(relocated); err != nil {
		t.Fatalf("Unmarshal failed on a relocated data block: %v", err)
	}
	if !bytes.Equal([]byte(decoded.Data), payload) {
		t.Errorf("the decoded bytes are not the relocated ones (%d of %d bytes)",
			len(decoded.Data), len(payload))
	}
}

func TestReadAndxResponseAbove64KiBRoundTrips(t *testing.T) {
	payload := countingPayload(0x10000 + 55)

	response := NewReadAndxResponse()
	response.Data = []types.UCHAR(payload)

	raw, err := response.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if want := types.USHORT(len(payload) & 0xFFFF); response.DataLength != want {
		t.Errorf("DataLength is 0x%04X, want 0x%04X", response.DataLength, want)
	}
	if want := types.USHORT(len(payload) >> 16); response.DataLengthHigh != want {
		t.Errorf("DataLengthHigh is 0x%04X, want 0x%04X", response.DataLengthHigh, want)
	}

	decoded := NewReadAndxResponse()
	if _, err := decoded.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if !bytes.Equal([]byte(decoded.Data), payload) {
		t.Errorf("the decoded bytes are not the ones read (%d of %d bytes)",
			len(decoded.Data), len(payload))
	}
}

func TestReadAndxResponseBatchedResolvesItsDataOffset(t *testing.T) {
	const chainOffset = 39
	payload := countingPayload(0x10000)

	response := NewReadAndxResponse()
	response.Data = []types.UCHAR(payload)
	response.SetChainOffset(chainOffset)

	raw, err := response.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if want := types.USHORT(readAndxResponseDataOffset + chainOffset); response.DataOffset != want {
		t.Errorf("DataOffset is %d, want %d", response.DataOffset, want)
	}

	decoded := NewReadAndxResponse()
	decoded.SetChainOffset(chainOffset)
	if _, err := decoded.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if !bytes.Equal([]byte(decoded.Data), payload) {
		t.Error("the decoded bytes are not the ones read")
	}
}

func TestReadAndxResponseWithoutADataOffsetReadsTheBlockTail(t *testing.T) {
	// A sender that leaves DataOffset unset is still understood: the bytes are
	// then the tail of the data block, which is where an unrelocated read puts
	// them. Refusing a reply this decoder can read would be the worse answer.
	payload := countingPayload(4096)

	response := NewReadAndxResponse()
	response.Data = []types.UCHAR(payload)

	raw, err := response.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	binary.LittleEndian.PutUint16(raw[readAndxResponseDataOffsetAt:readAndxResponseDataOffsetAt+2], 0)

	decoded := NewReadAndxResponse()
	if _, err := decoded.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if !bytes.Equal([]byte(decoded.Data), payload) {
		t.Error("the decoded bytes are not the ones read")
	}
}

// readAndxResponseDataOffsetAt is where the DataOffset field sits in a read
// response's own block: WordCount(1), the AndX block(4), Available(2),
// DataCompactionMode(2), Reserved1(2), DataLength(2).
const readAndxResponseDataOffsetAt = 1 + 4 + 2 + 2 + 2 + 2

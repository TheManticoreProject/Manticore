package message_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// The fields of a WriteAndx request that describe its data, in bytes from the
// start of the whole message: the SMB header, WordCount(1), the AndX block(4),
// FID(2), Offset(4), Timeout(4), WriteMode(2), Remaining(2), and then the three
// fields themselves. The commands package keeps these positions privately, so
// they are recomputed here from the field widths [MS-CIFS] section 2.2.4.43.1
// gives.
const (
	writeDataLengthHighAt = header.SMB_HEADER_SIZE + 1 + 4 + 2 + 4 + 4 + 2 + 2
	writeDataLengthAt     = writeDataLengthHighAt + 2
	writeDataOffsetAt     = writeDataLengthAt + 2
	writeByteCountAt      = writeDataOffsetAt + 2
)

// countingPayload is a block whose every byte identifies its own position.
func countingPayload(length int) []byte {
	block := make([]byte, length)
	for index := range block {
		block[index] = byte(index % 251)
	}
	return block
}

// TestUnmarshalFindsALargeWriteRelocatedPastAnAndXChain assembles the message
// [MS-CIFS] section 2.2.4.43.1 describes: an SMB_COM_WRITE_ANDX +
// SMB_COM_CLOSE chain whose write data has been relocated past the close, with
// ByteCount left at "1 + SMB_Parameters.Words.DataLength" — counting bytes that
// are not in the write's own block at all.
//
// It is the case that needs every part of this to work at once: the payload is
// too large for ByteCount to describe, it is nowhere near the write's data block,
// and the close that follows has to be decoded as well.
func TestUnmarshalFindsALargeWriteRelocatedPastAnAndXChain(t *testing.T) {
	payload := countingPayload(0x10000 + 21)

	// Build the chain with no data on the write, so its block ends at its pad
	// byte and the close follows immediately.
	msg := message.NewMessage()
	write := commands.NewWriteAndxRequest()
	write.FID = types.USHORT(0x00AB)
	write.Offset = types.ULONG(8192)
	msg.AddCommand(write)

	closeCommand := commands.NewCloseRequest()
	closeCommand.FID = types.USHORT(0x00AB)
	msg.AddCommand(closeCommand)

	raw, err := msg.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the chain: %v", err)
	}

	// Two bytes of alignment padding, then the relocated data, exactly as the
	// specification's example lays it out.
	relocated := append(append([]byte{}, raw...), 0x00, 0x00)
	dataAt := len(relocated)
	relocated = append(relocated, payload...)

	binary.LittleEndian.PutUint16(relocated[writeDataLengthAt:writeDataLengthAt+2],
		uint16(len(payload)&0xFFFF))
	binary.LittleEndian.PutUint16(relocated[writeDataLengthHighAt:writeDataLengthHighAt+2],
		uint16(len(payload)>>16))
	binary.LittleEndian.PutUint16(relocated[writeDataOffsetAt:writeDataOffsetAt+2], uint16(dataAt))
	binary.LittleEndian.PutUint16(relocated[writeByteCountAt:writeByteCountAt+2],
		uint16((1+len(payload))&0xFFFF))

	decoded := message.NewMessage()
	if err := decoded.Unmarshal(relocated); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	decodedWrite, ok := decoded.Command.(*commands.WriteAndxRequest)
	if !ok {
		t.Fatalf("the first command decoded as %T, want a WriteAndxRequest", decoded.Command)
	}
	if decodedWrite.FID != 0x00AB {
		t.Errorf("the write's FID is 0x%04X, want 0x00AB", decodedWrite.FID)
	}
	if len(decodedWrite.Data) != len(payload) {
		t.Fatalf("the write carries %d bytes, want the %d relocated", len(decodedWrite.Data), len(payload))
	}
	if !bytes.Equal([]byte(decodedWrite.Data), payload) {
		t.Error("the write's bytes are not the relocated ones")
	}

	// And the command batched behind it is still found, since the chain is
	// followed by AndXOffset and not by where the data happens to be.
	next := decodedWrite.GetNextCommand()
	if next == nil {
		t.Fatal("the close batched behind the write was not decoded")
	}
	decodedClose, ok := next.(*commands.CloseRequest)
	if !ok {
		t.Fatalf("the second command decoded as %T, want a CloseRequest", next)
	}
	if decodedClose.FID != 0x00AB {
		t.Errorf("the close's FID is 0x%04X, want 0x00AB", decodedClose.FID)
	}
}

// TestMarshalPositionsALargeWriteBatchedSecond asserts a write batched behind
// another command describes its data by where the data actually is.
//
// DataOffset is measured from the SMB header "regardless of the command
// request's position in an AndX chain", so a write that assumed it was first
// would point its server at the bytes of whatever preceded it.
func TestMarshalPositionsALargeWriteBatchedSecond(t *testing.T) {
	payload := countingPayload(0x10000 + 5)

	msg := message.NewMessage()

	// A lock taken and then written through is a real chain, and the lock request
	// carries no strings, so what this test asserts stays about the positioning.
	lead := commands.NewLockingAndxRequest()
	lead.FID = types.USHORT(0x0101)
	msg.AddCommand(lead)

	write := commands.NewWriteAndxRequest()
	write.FID = types.USHORT(0x0101)
	write.Data = []types.UCHAR(payload)
	msg.AddCommand(write)

	raw, err := msg.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the chain: %v", err)
	}

	decoded := message.NewMessage()
	if err := decoded.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	next := decoded.Command.GetNextCommand()
	if next == nil {
		t.Fatal("the write batched behind the open was not decoded")
	}
	decodedWrite, ok := next.(*commands.WriteAndxRequest)
	if !ok {
		t.Fatalf("the second command decoded as %T, want a WriteAndxRequest", next)
	}

	// The offset it wrote has to name the bytes it meant, in the whole message.
	at := int(decodedWrite.DataOffset)
	if at+len(payload) > len(raw) {
		t.Fatalf("DataOffset %d plus %d bytes runs past the %d-byte message", at, len(payload), len(raw))
	}
	if !bytes.Equal(raw[at:at+len(payload)], payload) {
		t.Error("DataOffset does not point at the bytes the write carried")
	}
	if !bytes.Equal([]byte(decodedWrite.Data), payload) {
		t.Errorf("the decoded write carries %d bytes that are not the ones written", len(decodedWrite.Data))
	}
}

package header_test

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
)

func TestNewHeaderDefaults(t *testing.T) {
	h := header.NewHeader()

	if !h.HasValidProtocolId() {
		t.Errorf("NewHeader ProtocolId = % x, want FE 53 4D 42", h.ProtocolId)
	}
	if h.StructureSize != header.SMB2_HEADER_STRUCTURE_SIZE {
		t.Errorf("NewHeader StructureSize = %d, want %d", h.StructureSize, header.SMB2_HEADER_STRUCTURE_SIZE)
	}
}

func TestHeaderMarshalLengthAndOffsets(t *testing.T) {
	h := header.NewHeader()
	h.CreditCharge = 0x0102
	h.Status = 0x04030201
	h.Command = codes.SMB2_CREATE
	h.Credit = 0x0807
	h.Flags = flags.SMB2_FLAGS_SIGNED
	h.NextCommand = 0x14131211
	h.MessageId = 0x1122334455667788
	h.Reserved = 0xAABBCCDD
	h.TreeId = 0x21222324
	h.SessionId = 0x99AABBCCDDEEFF00
	for i := range h.Signature {
		h.Signature[i] = byte(i)
	}

	buf, err := h.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}
	if len(buf) != header.SMB2_HEADER_SIZE {
		t.Fatalf("Marshal() produced %d bytes, want %d", len(buf), header.SMB2_HEADER_SIZE)
	}

	// Verify field placement at the spec offsets.
	if buf[0] != 0xFE || buf[1] != 'S' || buf[2] != 'M' || buf[3] != 'B' {
		t.Errorf("ProtocolId at 0..3 = % x, want FE 53 4D 42", buf[0:4])
	}
	if got := binary.LittleEndian.Uint16(buf[4:6]); got != 64 {
		t.Errorf("StructureSize at 4 = %d, want 64", got)
	}
	if got := binary.LittleEndian.Uint16(buf[6:8]); got != 0x0102 {
		t.Errorf("CreditCharge at 6 = 0x%04x, want 0x0102", got)
	}
	if got := binary.LittleEndian.Uint32(buf[8:12]); got != 0x04030201 {
		t.Errorf("Status at 8 = 0x%08x, want 0x04030201", got)
	}
	if got := binary.LittleEndian.Uint16(buf[12:14]); got != uint16(codes.SMB2_CREATE) {
		t.Errorf("Command at 12 = 0x%04x, want 0x%04x", got, uint16(codes.SMB2_CREATE))
	}
	if got := binary.LittleEndian.Uint16(buf[14:16]); got != 0x0807 {
		t.Errorf("Credit at 14 = 0x%04x, want 0x0807", got)
	}
	if got := binary.LittleEndian.Uint32(buf[16:20]); got != uint32(flags.SMB2_FLAGS_SIGNED) {
		t.Errorf("Flags at 16 = 0x%08x, want 0x%08x", got, uint32(flags.SMB2_FLAGS_SIGNED))
	}
	if got := binary.LittleEndian.Uint32(buf[20:24]); got != 0x14131211 {
		t.Errorf("NextCommand at 20 = 0x%08x, want 0x14131211", got)
	}
	if got := binary.LittleEndian.Uint64(buf[24:32]); got != 0x1122334455667788 {
		t.Errorf("MessageId at 24 = 0x%016x, want 0x1122334455667788", got)
	}
	if got := binary.LittleEndian.Uint32(buf[32:36]); got != 0xAABBCCDD {
		t.Errorf("Reserved at 32 = 0x%08x, want 0xAABBCCDD", got)
	}
	if got := binary.LittleEndian.Uint32(buf[36:40]); got != 0x21222324 {
		t.Errorf("TreeId at 36 = 0x%08x, want 0x21222324", got)
	}
	if got := binary.LittleEndian.Uint64(buf[40:48]); got != 0x99AABBCCDDEEFF00 {
		t.Errorf("SessionId at 40 = 0x%016x, want 0x99AABBCCDDEEFF00", got)
	}
	for i := 0; i < header.SMB2_SIGNATURE_SIZE; i++ {
		if buf[48+i] != byte(i) {
			t.Errorf("Signature byte %d at %d = 0x%02x, want 0x%02x", i, 48+i, buf[48+i], byte(i))
		}
	}
}

func TestHeaderSyncRoundTrip(t *testing.T) {
	original := header.NewHeader()
	original.CreditCharge = 1
	original.Status = 0xC000000D
	original.Command = codes.SMB2_TREE_CONNECT
	original.Credit = 2
	original.Flags = flags.SMB2_FLAGS_SERVER_TO_REDIR
	original.NextCommand = 0
	original.MessageId = 0x00000000DEADBEEF
	original.Reserved = 0
	original.TreeId = 0x00000005
	original.SessionId = 0x0000000012345678

	buf, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var decoded header.Header
	n, err := decoded.Unmarshal(buf)
	if err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if n != header.SMB2_HEADER_SIZE {
		t.Errorf("Unmarshal() read %d bytes, want %d", n, header.SMB2_HEADER_SIZE)
	}
	if decoded != *original {
		t.Errorf("SYNC round-trip mismatch:\n got  %+v\n want %+v", decoded, *original)
	}
	if decoded.IsAsync() {
		t.Errorf("decoded header should not be ASYNC")
	}
	if !decoded.IsResponse() {
		t.Errorf("decoded header should be a response (SERVER_TO_REDIR set)")
	}
}

func TestHeaderAsyncRoundTrip(t *testing.T) {
	original := header.NewHeader()
	original.Command = codes.SMB2_READ
	original.Flags = flags.SMB2_FLAGS_ASYNC_COMMAND | flags.SMB2_FLAGS_SERVER_TO_REDIR
	original.MessageId = 0x0102030405060708
	original.AsyncId = 0x1122334455667788
	original.SessionId = 0xAABBCCDDEEFF0011

	buf, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	// In the ASYNC form, AsyncId occupies offsets 32..39.
	if got := binary.LittleEndian.Uint64(buf[32:40]); got != original.AsyncId {
		t.Errorf("AsyncId at 32 = 0x%016x, want 0x%016x", got, original.AsyncId)
	}

	var decoded header.Header
	if _, err := decoded.Unmarshal(buf); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if !decoded.IsAsync() {
		t.Fatalf("decoded header should be ASYNC")
	}
	if decoded.AsyncId != original.AsyncId {
		t.Errorf("AsyncId = 0x%016x, want 0x%016x", decoded.AsyncId, original.AsyncId)
	}
	// SYNC-only views must be zeroed when the ASYNC bit is set.
	if decoded.Reserved != 0 || decoded.TreeId != 0 {
		t.Errorf("ASYNC decode left Reserved=0x%08x TreeId=0x%08x, want both 0", decoded.Reserved, decoded.TreeId)
	}
	if decoded != *original {
		t.Errorf("ASYNC round-trip mismatch:\n got  %+v\n want %+v", decoded, *original)
	}
}

func TestHeaderUnmarshalShort(t *testing.T) {
	var h header.Header
	if _, err := h.Unmarshal(make([]byte, header.SMB2_HEADER_SIZE-1)); err == nil {
		t.Errorf("expected error unmarshalling a buffer shorter than %d bytes", header.SMB2_HEADER_SIZE)
	}
}

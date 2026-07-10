package message_test

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/llmnr/class"
	"github.com/TheManticoreProject/Manticore/network/llmnr/constants"
	"github.com/TheManticoreProject/Manticore/network/llmnr/domain_name"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
	"github.com/TheManticoreProject/Manticore/network/llmnr/resourcerecord"
)

// TestWriteReadTCPMessageRoundTrip confirms a payload framed by WriteTCPMessage
// is recovered byte-for-byte by ReadTCPMessage, and that the on-wire framing is
// the RFC 1035 §4.2.2 two-byte big-endian length prefix (excluding the prefix
// itself) followed by the payload.
func TestWriteReadTCPMessageRoundTrip(t *testing.T) {
	payload := []byte{0x00, 0x01, 0x02, 0x03, 0xAA, 0xBB}

	var buf bytes.Buffer
	if err := message.WriteTCPMessage(&buf, payload); err != nil {
		t.Fatalf("WriteTCPMessage() error = %v", err)
	}

	framed := buf.Bytes()
	if len(framed) != len(payload)+2 {
		t.Fatalf("framed length = %d, want %d (payload + 2-byte prefix)", len(framed), len(payload)+2)
	}
	if got := binary.BigEndian.Uint16(framed[:2]); int(got) != len(payload) {
		t.Errorf("length prefix = %d, want %d (excludes the prefix)", got, len(payload))
	}
	if !bytes.Equal(framed[2:], payload) {
		t.Errorf("framed body = % x, want % x", framed[2:], payload)
	}

	got, err := message.ReadTCPMessage(&buf)
	if err != nil {
		t.Fatalf("ReadTCPMessage() error = %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("ReadTCPMessage() = % x, want % x", got, payload)
	}
}

// TestReadTCPMessageShortPrefix confirms a stream carrying fewer than the two
// prefix bytes is reported as an error rather than silently succeeding.
func TestReadTCPMessageShortPrefix(t *testing.T) {
	if _, err := message.ReadTCPMessage(bytes.NewReader([]byte{0x00})); err == nil {
		t.Error("ReadTCPMessage() error = nil, want error on a truncated length prefix")
	}
}

// TestReadTCPMessageShortBody confirms a stream whose length prefix announces
// more bytes than are actually present is reported as an error (a short read on
// the body), rather than returning a partial message.
func TestReadTCPMessageShortBody(t *testing.T) {
	// Prefix says 4 bytes, but only 2 follow.
	stream := []byte{0x00, 0x04, 0xDE, 0xAD}
	_, err := message.ReadTCPMessage(bytes.NewReader(stream))
	if err == nil {
		t.Fatal("ReadTCPMessage() error = nil, want error on a truncated body")
	}
	// io.ReadFull surfaces an unexpected EOF for a partial body.
	if err != io.ErrUnexpectedEOF && !strings.Contains(err.Error(), io.ErrUnexpectedEOF.Error()) {
		t.Errorf("ReadTCPMessage() error = %v, want it to wrap %v", err, io.ErrUnexpectedEOF)
	}
}

// TestReadTCPMessageZeroLength confirms a zero-length prefix yields an empty,
// non-nil payload and no error (a valid, if empty, frame).
func TestReadTCPMessageZeroLength(t *testing.T) {
	got, err := message.ReadTCPMessage(bytes.NewReader([]byte{0x00, 0x00}))
	if err != nil {
		t.Fatalf("ReadTCPMessage() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ReadTCPMessage() = % x, want empty", got)
	}
}

// TestMarshalWithTruncationFits confirms that a small message which fits within
// the limit is returned unchanged, with truncated == false and the TC bit clear.
func TestMarshalWithTruncationFits(t *testing.T) {
	m := message.NewMessage()
	m.SetResponse()
	if err := m.AddAnswerClassINTypeA("host.local", "10.7.0.10"); err != nil {
		t.Fatalf("AddAnswerClassINTypeA() error = %v", err)
	}

	encoded, truncated, err := m.MarshalWithTruncation(constants.MaxPacketSize)
	if err != nil {
		t.Fatalf("MarshalWithTruncation() error = %v", err)
	}
	if truncated {
		t.Error("MarshalWithTruncation() truncated = true, want false for a small message")
	}

	full, err := m.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if !bytes.Equal(encoded, full) {
		t.Error("MarshalWithTruncation() changed the encoding of a message that fits")
	}

	decoded := &message.Message{}
	if _, err := decoded.Unmarshal(encoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if decoded.Header.Flags.IsTruncation() {
		t.Error("TC bit set on a message that fits within the limit")
	}
}

// TestMarshalWithTruncationSetsTC builds a response with enough answer records to
// exceed the 512-byte UDP limit, then confirms MarshalWithTruncation sets the TC
// bit, drops records so the result fits within the limit, and still decodes as a
// well-formed message (its header counts matching the records it retained).
func TestMarshalWithTruncationSetsTC(t *testing.T) {
	m := message.NewMessage()
	m.SetResponse()
	if err := m.AddQuestion("host.local", llmnr_type.TypeA, class.ClassIN); err != nil {
		t.Fatalf("AddQuestion() error = %v", err)
	}

	// Each answer carries a distinct ~30-byte owner name plus a 4-byte A record;
	// 40 of them comfortably exceed the 512-byte UDP datagram limit. Names are
	// not compressed on marshal, so this reliably overflows. Answers are added
	// directly (rather than via AddAnswerClassINTypeA, which would also grow the
	// question section) so the single question is the only one present.
	for i := 0; i < 40; i++ {
		name := strings.Repeat("a", 20) + "-host" + string(rune('a'+i%26)) + ".local"
		rr := resourcerecord.ResourceRecord{
			Name:  domain_name.DomainName(name),
			Type:  llmnr_type.TypeA,
			Class: class.ClassIN,
			TTL:   30,
			RData: resourcerecord.IPv4ToRData("10.7.0.10"),
		}
		if err := m.AddAnswer(rr); err != nil {
			t.Fatalf("AddAnswer(%q) error = %v", name, err)
		}
	}

	full, err := m.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if len(full) <= constants.MaxPacketSize {
		t.Fatalf("test fixture only %d bytes, want > %d so truncation is exercised", len(full), constants.MaxPacketSize)
	}

	encoded, truncated, err := m.MarshalWithTruncation(constants.MaxPacketSize)
	if err != nil {
		t.Fatalf("MarshalWithTruncation() error = %v", err)
	}
	if !truncated {
		t.Fatal("MarshalWithTruncation() truncated = false, want true for an oversized message")
	}
	if len(encoded) > constants.MaxPacketSize {
		t.Errorf("truncated encoding = %d bytes, want <= %d", len(encoded), constants.MaxPacketSize)
	}

	// The receiver's message must be left intact (truncation applied to a copy).
	if m.Header.Flags.IsTruncation() {
		t.Error("MarshalWithTruncation() mutated the receiver's TC bit")
	}
	if len(m.Answers) != 40 {
		t.Errorf("MarshalWithTruncation() mutated the receiver's answers: got %d, want 40", len(m.Answers))
	}

	decoded := &message.Message{}
	if _, err := decoded.Unmarshal(encoded); err != nil {
		t.Fatalf("Unmarshal(truncated) error = %v", err)
	}
	if !decoded.Header.Flags.IsTruncation() {
		t.Error("truncated response does not have the TC bit set")
	}
	if len(decoded.Answers) >= 40 {
		t.Errorf("truncated response kept %d answers, want fewer than the original 40", len(decoded.Answers))
	}
	if err := decoded.Validate(); err != nil {
		t.Errorf("truncated response is not well-formed: %v", err)
	}
}

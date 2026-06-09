package message_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
)

// fakeCommand is a test-only command body used to exercise the message envelope
// and compounding before any concrete SMB2 command structure exists. The
// production dispatcher has no generic command (matching SMB 1.0), so the
// happy-path Unmarshal/UnmarshalCompound round-trips — which decode through the
// dispatcher — are covered once the first concrete command lands (Phase 3).
type fakeCommand struct {
	command_interface.Command
	payload []byte
}

func newFakeCommand(code codes.CommandCode, payload []byte) *fakeCommand {
	c := &fakeCommand{payload: payload}
	c.SetCommandCode(code)
	return c
}

func (c *fakeCommand) Marshal() ([]byte, error)        { return c.payload, nil }
func (c *fakeCommand) Unmarshal(d []byte) (int, error) { c.payload = d; return len(d), nil }

func newFakeMessage(code codes.CommandCode, messageId uint64, payload []byte) *message.Message {
	m := message.NewMessage()
	m.Header.MessageId = messageId
	m.SetCommand(newFakeCommand(code, payload))
	return m
}

func TestMessageMarshalSingle(t *testing.T) {
	body := []byte{0x09, 0x00, 0x00, 0x00, 0xDE, 0xAD}
	m := newFakeMessage(codes.SMB2_TREE_CONNECT, 0x42, body)

	wire, err := m.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != header.SMB2_HEADER_SIZE+len(body) {
		t.Fatalf("wire length = %d, want %d", len(wire), header.SMB2_HEADER_SIZE+len(body))
	}
	// SetCommand must copy the command code into the header.
	if got := binary.LittleEndian.Uint16(wire[12:14]); got != uint16(codes.SMB2_TREE_CONNECT) {
		t.Errorf("header Command = 0x%04x, want TREE_CONNECT", got)
	}
	// NextCommand of a single message is 0.
	if got := binary.LittleEndian.Uint32(wire[20:24]); got != 0 {
		t.Errorf("NextCommand = %d, want 0 for a single message", got)
	}
	if !bytes.Equal(wire[header.SMB2_HEADER_SIZE:], body) {
		t.Errorf("body = % x, want % x", wire[header.SMB2_HEADER_SIZE:], body)
	}
}

// TestMessageRoundTripWithRealCommand exercises the full envelope decode path
// (Message.Unmarshal -> dispatcher -> concrete command) now that a concrete
// command exists, the way the SMB 1.0 message tests use real commands.
func TestMessageRoundTripWithRealCommand(t *testing.T) {
	req := commands.NewNegotiateRequest()
	req.AddDialect(dialects.SMB2_DIALECT_2_0_2)

	m := message.NewMessage()
	m.Header.MessageId = 7
	m.SetCommand(req)

	wire, err := m.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	decoded := message.NewMessage()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Header.Command != codes.SMB2_NEGOTIATE || decoded.Header.MessageId != 7 {
		t.Errorf("header mismatch: %+v", decoded.Header)
	}
	neg, ok := decoded.Command.(*commands.NegotiateRequest)
	if !ok {
		t.Fatalf("decoded command is %T, want *NegotiateRequest", decoded.Command)
	}
	if len(neg.Dialects) != 1 || neg.Dialects[0] != dialects.SMB2_DIALECT_2_0_2 {
		t.Errorf("dialects mismatch: %v", neg.Dialects)
	}

	// A response is dispatched to the response-side type via SERVER_TO_REDIR.
	resp := commands.NewNegotiateResponse()
	resp.DialectRevision = dialects.SMB2_DIALECT_2_0_2
	rm := message.NewMessage()
	rm.Header.AddFlags(flags.SMB2_FLAGS_SERVER_TO_REDIR)
	rm.SetCommand(resp)
	rwire, err := rm.Marshal()
	if err != nil {
		t.Fatalf("response Marshal: %v", err)
	}
	rdecoded := message.NewMessage()
	if _, err := rdecoded.Unmarshal(rwire); err != nil {
		t.Fatalf("response Unmarshal: %v", err)
	}
	if _, ok := rdecoded.Command.(*commands.NegotiateResponse); !ok {
		t.Errorf("decoded response command is %T, want *NegotiateResponse", rdecoded.Command)
	}
}

func TestMessageMarshalNoCommand(t *testing.T) {
	m := message.NewMessage()
	if _, err := m.Marshal(); err == nil {
		t.Errorf("expected error marshalling a message with no command")
	}
}

func TestMarshalCompoundLayout(t *testing.T) {
	// 4-byte body makes segment 0 occupy 68 bytes, which must be padded to 72
	// (the next 8-byte boundary) so segment 1's header starts 8-aligned.
	body0 := []byte{0x09, 0x00, 0x00, 0x00}
	body1 := []byte{0x18, 0x00, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55}
	msgs := []*message.Message{
		newFakeMessage(codes.SMB2_CREATE, 1, body0),
		newFakeMessage(codes.SMB2_CLOSE, 2, body1),
	}

	wire, err := message.MarshalCompound(msgs)
	if err != nil {
		t.Fatalf("MarshalCompound: %v", err)
	}

	// Segment 0 header: Command=CREATE, NextCommand=72 (8-aligned).
	if got := binary.LittleEndian.Uint16(wire[12:14]); got != uint16(codes.SMB2_CREATE) {
		t.Errorf("segment 0 Command = 0x%04x, want CREATE", got)
	}
	if got := binary.LittleEndian.Uint32(wire[20:24]); got != 72 {
		t.Errorf("segment 0 NextCommand = %d, want 72", got)
	}
	// The 4 bytes between body0 end (68) and the next header (72) must be zero pad.
	if !bytes.Equal(wire[68:72], []byte{0, 0, 0, 0}) {
		t.Errorf("padding bytes = % x, want zeros", wire[68:72])
	}
	// Segment 1 header starts at offset 72: Command=CLOSE, NextCommand=0 (last).
	if got := binary.LittleEndian.Uint16(wire[72+12 : 72+14]); got != uint16(codes.SMB2_CLOSE) {
		t.Errorf("segment 1 Command = 0x%04x, want CLOSE", got)
	}
	if got := binary.LittleEndian.Uint32(wire[72+20 : 72+24]); got != 0 {
		t.Errorf("segment 1 NextCommand = %d, want 0 (last)", got)
	}
	// Total = 72 (padded segment 0) + 64 (header) + 8 (body1).
	if len(wire) != 72+header.SMB2_HEADER_SIZE+len(body1) {
		t.Errorf("total wire length = %d, want %d", len(wire), 72+header.SMB2_HEADER_SIZE+len(body1))
	}
}

func TestMarshalCompoundEmpty(t *testing.T) {
	if _, err := message.MarshalCompound(nil); err == nil {
		t.Errorf("expected error marshalling an empty compound")
	}
}

func TestMessageUnmarshalShort(t *testing.T) {
	m := message.NewMessage()
	if _, err := m.Unmarshal(make([]byte, header.SMB2_HEADER_SIZE-1)); err == nil {
		t.Errorf("expected error unmarshalling a buffer shorter than the header")
	}
}

func TestMessageUnmarshalBadNextCommand(t *testing.T) {
	// The NextCommand bounds check runs before the command dispatcher, so this
	// error path is testable without a concrete command.
	m := newFakeMessage(codes.SMB2_READ, 1, []byte{0x31, 0x00})
	wire, err := m.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	binary.LittleEndian.PutUint32(wire[20:24], 0xFFFF) // out-of-bounds offset

	decoded := message.NewMessage()
	if _, err := decoded.Unmarshal(wire); err == nil {
		t.Errorf("expected error on out-of-bounds NextCommand")
	}
}

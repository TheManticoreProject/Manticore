package client_test

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// capturingTransport records the last Send payload and replays a canned response.
type capturingTransport struct {
	sent     []byte
	response []byte
}

func (m *capturingTransport) Connect(ipaddr net.IP, port int) error { return nil }
func (m *capturingTransport) Close() error                          { return nil }
func (m *capturingTransport) Send(data []byte) (int, error) {
	m.sent = append([]byte(nil), data...)
	return len(data), nil
}
func (m *capturingTransport) Receive() ([]byte, error) { return m.response, nil }
func (m *capturingTransport) IsConnected() bool        { return true }
func (m *capturingTransport) SetTimeout(time.Duration) {}

func newSessionClient(tr *capturingTransport) *client.Client {
	c := &client.Client{
		Transport:  tr,
		Connection: &client.Connection{Server: &client.Server{}},
	}
	c.Session = &client.Session{SessionUID: 1, TreeID: 1}
	return c
}

func marshalResponse(t *testing.T, cmd command_interface.CommandInterface) []byte {
	t.Helper()
	msg := message.NewMessage()
	msg.AddCommand(cmd) // also sets Header.Command from the command code
	msg.Header.SetFlags(flags.FLAGS_REPLY)
	raw, err := msg.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal canned response: %v", err)
	}
	return raw
}

// TestWriteFileLargeOffsetCarriesOffsetHigh verifies that WriteFile preserves the
// full 64-bit file offset (OffsetHigh) for writes past 4 GiB, and that the request
// DataOffset points at the data block actually emitted.
func TestWriteFileLargeOffsetCarriesOffsetHigh(t *testing.T) {
	offset := uint64(0x1_0000_0010) // 4 GiB + 16: high dword = 1, low dword = 0x10
	data := []byte("payload-bytes")

	resp := commands.NewWriteAndxResponse()
	resp.Count = types.USHORT(len(data))

	tr := &capturingTransport{response: marshalResponse(t, resp)}
	c := newSessionClient(tr)

	n, err := c.WriteFile(client.FID(0x4242), offset, data)
	if err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if n != len(data) {
		t.Fatalf("WriteFile wrote %d bytes, want %d", n, len(data))
	}

	// Parse the request the client actually sent.
	reqMsg := message.NewMessage()
	if err := reqMsg.Unmarshal(tr.sent); err != nil {
		t.Fatalf("failed to unmarshal captured request: %v", err)
	}
	req, ok := reqMsg.Command.(*commands.WriteAndxRequest)
	if !ok {
		t.Fatalf("captured command is %T, want *WriteAndxRequest", reqMsg.Command)
	}

	if uint32(req.Offset) != uint32(offset) {
		t.Errorf("Offset = 0x%08x, want 0x%08x", uint32(req.Offset), uint32(offset))
	}
	if uint32(req.OffsetHigh) != uint32(offset>>32) {
		t.Errorf("OffsetHigh = 0x%08x, want 0x%08x (offset was truncated to 32 bits)", uint32(req.OffsetHigh), uint32(offset>>32))
	}

	// The data must actually live at the byte position the request advertises.
	off := int(req.DataOffset)
	if off+len(data) > len(tr.sent) {
		t.Fatalf("DataOffset %d + len %d exceeds request size %d", off, len(data), len(tr.sent))
	}
	if !bytes.Equal(tr.sent[off:off+len(data)], data) {
		t.Errorf("data at DataOffset %d = %q, want %q", off, tr.sent[off:off+len(data)], data)
	}
}

// TestReadFileLargeOffsetCarriesOffsetHigh verifies that ReadFile preserves the
// full 64-bit file offset (OffsetHigh) for reads past 4 GiB.
func TestReadFileLargeOffsetCarriesOffsetHigh(t *testing.T) {
	offset := uint64(0x2_0000_0020) // 8 GiB + 32: high dword = 2, low dword = 0x20

	// A response with DataLength 0 signals end-of-file, so ReadFile issues exactly
	// one request and returns.
	resp := commands.NewReadAndxResponse()
	resp.DataLength = types.USHORT(0)

	tr := &capturingTransport{response: marshalResponse(t, resp)}
	c := newSessionClient(tr)

	if _, err := c.ReadFile(client.FID(0x4242), offset, 64); err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	reqMsg := message.NewMessage()
	if err := reqMsg.Unmarshal(tr.sent); err != nil {
		t.Fatalf("failed to unmarshal captured request: %v", err)
	}
	req, ok := reqMsg.Command.(*commands.ReadAndxRequest)
	if !ok {
		t.Fatalf("captured command is %T, want *ReadAndxRequest", reqMsg.Command)
	}

	if uint32(req.Offset) != uint32(offset) {
		t.Errorf("Offset = 0x%08x, want 0x%08x", uint32(req.Offset), uint32(offset))
	}
	if uint32(req.OffsetHigh) != uint32(offset>>32) {
		t.Errorf("OffsetHigh = 0x%08x, want 0x%08x (offset was truncated to 32 bits)", uint32(req.OffsetHigh), uint32(offset>>32))
	}
}

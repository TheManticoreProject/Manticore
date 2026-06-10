package client

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// recordingTransport records every Send payload and replays a canned response.
type recordingTransport struct {
	sent     [][]byte
	response []byte
}

func (m *recordingTransport) Connect(ipaddr net.IP, port int) error { return nil }
func (m *recordingTransport) Close() error                          { return nil }
func (m *recordingTransport) Send(data []byte) (int, error) {
	m.sent = append(m.sent, append([]byte(nil), data...))
	return len(data), nil
}
func (m *recordingTransport) Receive() ([]byte, error) { return m.response, nil }
func (m *recordingTransport) IsConnected() bool        { return true }
func (m *recordingTransport) SetTimeout(time.Duration) {}

func TestPlanTransaction2Chunks(t *testing.T) {
	// Reassembling the chunk runs by displacement must reproduce the inputs, every
	// chunk must stay within the payload budget, and data must not appear before all
	// parameters are placed.
	cases := []struct {
		name           string
		params, data   []byte
		maxPayload     int
		wantChunks     int
	}{
		{"empty", nil, nil, 16, 1},
		{"fits in one", bytes.Repeat([]byte{1}, 10), bytes.Repeat([]byte{2}, 10), 64, 1},
		{"params only split", bytes.Repeat([]byte{3}, 100), nil, 40, 3},
		{"params then data", bytes.Repeat([]byte{4}, 100), bytes.Repeat([]byte{5}, 300), 64, 7},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chunks := planTransaction2Chunks(tc.params, tc.data, tc.maxPayload)
			if len(chunks) != tc.wantChunks {
				t.Errorf("got %d chunks, want %d", len(chunks), tc.wantChunks)
			}

			gotParams := make([]byte, 0, len(tc.params))
			gotData := make([]byte, 0, len(tc.data))
			for i, ch := range chunks {
				if len(ch.params)+len(ch.data) > tc.maxPayload {
					t.Errorf("chunk %d payload %d exceeds budget %d", i, len(ch.params)+len(ch.data), tc.maxPayload)
				}
				if ch.parameterDisplacement != len(gotParams) {
					t.Errorf("chunk %d parameterDisplacement %d, want %d", i, ch.parameterDisplacement, len(gotParams))
				}
				if ch.dataDisplacement != len(gotData) {
					t.Errorf("chunk %d dataDisplacement %d, want %d", i, ch.dataDisplacement, len(gotData))
				}
				// Data may only appear once every parameter byte has been placed,
				// counting this chunk's own parameters.
				if len(ch.data) > 0 && len(gotParams)+len(ch.params) < len(tc.params) {
					t.Errorf("chunk %d carries data before all parameters are placed", i)
				}
				gotParams = append(gotParams, ch.params...)
				gotData = append(gotData, ch.data...)
			}
			if !bytes.Equal(gotParams, tc.params) {
				t.Errorf("reassembled params mismatch")
			}
			if !bytes.Equal(gotData, tc.data) {
				t.Errorf("reassembled data mismatch")
			}
		})
	}
}

func TestTrans2FragmentsLargeRequest(t *testing.T) {
	params := make([]byte, 100)
	for i := range params {
		params[i] = byte(i)
	}
	data := make([]byte, 300)
	for i := range data {
		data[i] = byte(i ^ 0xAA)
	}

	// Canned response: a single TRANS2 reply with a small data block (DataOffset 55).
	respCmd := commands.NewTransaction2Response()
	respCmd.TotalDataCount = types.USHORT(2)
	respCmd.DataCount = types.USHORT(2)
	respCmd.DataOffset = types.USHORT(55)
	respCmd.Trans2_Data = []types.UCHAR{0xDE, 0xAD}
	respMsg := message.NewMessage()
	respMsg.Header.SetFlags(flags.FLAGS_REPLY)
	respMsg.AddCommand(respCmd)
	respRaw, err := respMsg.Marshal()
	if err != nil {
		t.Fatalf("marshal canned response: %v", err)
	}

	rec := &recordingTransport{response: respRaw}
	c := &Client{
		Transport:  rec,
		Connection: &Connection{Server: &Server{MaxBufferSize: 192}}, // payload budget = 192-128 = 64
		Session:    &Session{},
	}

	if _, _, err := c.trans2(0x0001, params, data); err != nil {
		t.Fatalf("trans2: %v", err)
	}

	if len(rec.sent) < 2 {
		t.Fatalf("expected the request to be fragmented into >= 2 messages, got %d", len(rec.sent))
	}

	// Reassemble the parameter and data runs from the emitted messages by
	// displacement and confirm they reproduce the original buffers.
	gotParams := make([]byte, len(params))
	gotData := make([]byte, len(data))
	for i, raw := range rec.sent {
		msg := message.NewMessage()
		if err := msg.Unmarshal(raw); err != nil {
			t.Fatalf("decode sent message %d: %v", i, err)
		}
		var p, d []byte
		var pDisp, dDisp int
		switch cmd := msg.Command.(type) {
		case *commands.Transaction2Request:
			if i != 0 {
				t.Errorf("message %d is the primary but is not first", i)
			}
			if int(cmd.TotalParameterCount) != len(params) || int(cmd.TotalDataCount) != len(data) {
				t.Errorf("primary totals = (%d,%d), want (%d,%d)", cmd.TotalParameterCount, cmd.TotalDataCount, len(params), len(data))
			}
			p, d = []byte(cmd.Trans2_Parameters), []byte(cmd.Trans2_Data)
			pDisp, dDisp = 0, 0
		case *commands.Transaction2SecondaryRequest:
			p, d = []byte(cmd.Trans2_Parameters), []byte(cmd.Trans2_Data)
			pDisp, dDisp = int(cmd.ParameterDisplacement), int(cmd.DataDisplacement)
		default:
			t.Fatalf("message %d has unexpected type %T", i, msg.Command)
		}
		copy(gotParams[pDisp:], p)
		copy(gotData[dDisp:], d)
	}

	if !bytes.Equal(gotParams, params) {
		t.Error("reassembled parameters do not match the original")
	}
	if !bytes.Equal(gotData, data) {
		t.Error("reassembled data does not match the original")
	}
}

func TestTrans2FragmentationRefusedWhileSigning(t *testing.T) {
	c := &Client{
		Transport: &recordingTransport{},
		Connection: &Connection{
			Server:          &Server{MaxBufferSize: 192},
			IsSigningActive: true,
		},
		Session: &Session{},
	}
	if _, _, err := c.trans2(0x0001, make([]byte, 100), make([]byte, 300)); err == nil {
		t.Error("expected an error when a fragmented Transaction2 is sent with signing active")
	}
}

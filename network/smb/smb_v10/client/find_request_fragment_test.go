package client

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
)

// --- planTrans2Send -------------------------------------------------------

// checkPlanCoversPayload asserts the plan is contiguous, covers exactly the
// totals, and places all parameter bytes before any data byte (per [MS-CIFS]).
func checkPlanCoversPayload(t *testing.T, plan []trans2SendChunk, totalParams, totalData int) {
	t.Helper()
	pExpect, dExpect := 0, 0
	paramsDone := false
	for i, c := range plan {
		if c.paramDisplacement != pExpect {
			t.Fatalf("chunk %d paramDisplacement = %d, want %d", i, c.paramDisplacement, pExpect)
		}
		if c.dataDisplacement != dExpect {
			t.Fatalf("chunk %d dataDisplacement = %d, want %d", i, c.dataDisplacement, dExpect)
		}
		if c.dataLen > 0 && !paramsDone && c.paramDisplacement+c.paramLen < totalParams {
			t.Fatalf("chunk %d carries data before all parameters were placed", i)
		}
		pExpect += c.paramLen
		dExpect += c.dataLen
		if pExpect >= totalParams {
			paramsDone = true
		}
	}
	if pExpect != totalParams {
		t.Errorf("plan covers %d parameter bytes, want %d", pExpect, totalParams)
	}
	if dExpect != totalData {
		t.Errorf("plan covers %d data bytes, want %d", dExpect, totalData)
	}
}

func TestPlanTrans2SendSingleMessage(t *testing.T) {
	plan := planTrans2Send(10, 20, 4096)
	if len(plan) != 1 {
		t.Fatalf("expected a single chunk, got %d", len(plan))
	}
	if plan[0] != (trans2SendChunk{0, 10, 0, 20}) {
		t.Errorf("chunk = %+v, want {0 10 0 20}", plan[0])
	}
}

func TestPlanTrans2SendFragments(t *testing.T) {
	cases := []struct{ params, data, buffer int }{
		{0, 500, 256},   // data only, must fragment
		{300, 0, 256},   // parameters only, must fragment
		{100, 400, 256}, // both, must fragment
		{50, 50, 4096},  // fits in one
	}
	for _, tc := range cases {
		plan := planTrans2Send(tc.params, tc.data, tc.buffer)
		if len(plan) == 0 {
			t.Fatalf("planTrans2Send(%d,%d,%d) returned no chunks", tc.params, tc.data, tc.buffer)
		}
		checkPlanCoversPayload(t, plan, tc.params, tc.data)
	}
}

// --- end-to-end send framing ---------------------------------------------

// recordingTransport records every Send and replays a single canned response.
type recordingTransport struct {
	sent     [][]byte
	response []byte
}

func (m *recordingTransport) Connect(net.IP, int) error { return nil }
func (m *recordingTransport) Close() error              { return nil }
func (m *recordingTransport) Send(d []byte) (int, error) {
	m.sent = append(m.sent, append([]byte(nil), d...))
	return len(d), nil
}
func (m *recordingTransport) Receive() ([]byte, error) { return m.response, nil }
func (m *recordingTransport) IsConnected() bool        { return true }

func le16(b []byte) int { return int(binary.LittleEndian.Uint16(b)) }

// TestTrans2RequestFragmentation drives trans2 with a payload too large for one
// SMB buffer and verifies that it is split into a primary SMB_COM_TRANSACTION2
// message followed by SMB_COM_TRANSACTION2_SECONDARY continuations, that every
// message advertises the full totals, and that the per-message runs reassemble
// (by displacement) into the original parameter and data buffers.
func TestTrans2RequestFragmentation(t *testing.T) {
	params := make([]byte, 100)
	for i := range params {
		params[i] = byte(i)
	}
	data := make([]byte, 400)
	for i := range data {
		data[i] = byte(i*7 + 1)
	}

	// A valid, empty Transaction2 response so trans2 completes after sending.
	cannedResp := makeTrans2ResponseMsg(0, nil, 0, 0, nil, 0)

	tr := &recordingTransport{response: cannedResp}
	c := &Client{
		Transport:  tr,
		Connection: &Connection{Server: &Server{MaxBufferSize: 256}},
		Session:    &Session{SessionUID: 1, TreeID: 1},
	}

	if _, _, err := c.trans2(0x0001, params, data); err != nil {
		t.Fatalf("trans2 error: %v", err)
	}

	if len(tr.sent) < 2 {
		t.Fatalf("expected the request to be fragmented into >= 2 messages, got %d", len(tr.sent))
	}

	gotParams := make([]byte, len(params))
	gotData := make([]byte, len(data))

	for i, msg := range tr.sent {
		cmd := msg[4] // SMB header: bytes 0-3 = \xffSMB, byte 4 = Command
		if i == 0 {
			if cmd != byte(codes.SMB_COM_TRANSACTION2) {
				t.Errorf("message 0 command = 0x%02x, want 0x%02x (TRANSACTION2)", cmd, byte(codes.SMB_COM_TRANSACTION2))
			}
			if le16(msg[33:35]) != len(params) || le16(msg[35:37]) != len(data) {
				t.Errorf("primary totals = (%d,%d), want (%d,%d)", le16(msg[33:35]), le16(msg[35:37]), len(params), len(data))
			}
			pc, po := le16(msg[51:53]), le16(msg[53:55])
			dc, do := le16(msg[55:57]), le16(msg[57:59])
			copy(gotParams[0:], msg[po:po+pc])
			if dc > 0 {
				copy(gotData[0:], msg[do:do+dc])
			}
		} else {
			if cmd != byte(codes.SMB_COM_TRANSACTION2_SECONDARY) {
				t.Errorf("message %d command = 0x%02x, want 0x%02x (TRANSACTION2_SECONDARY)", i, cmd, byte(codes.SMB_COM_TRANSACTION2_SECONDARY))
			}
			if le16(msg[33:35]) != len(params) || le16(msg[35:37]) != len(data) {
				t.Errorf("secondary %d totals = (%d,%d), want (%d,%d)", i, le16(msg[33:35]), le16(msg[35:37]), len(params), len(data))
			}
			pc, po, pd := le16(msg[37:39]), le16(msg[39:41]), le16(msg[41:43])
			dc, do, dd := le16(msg[43:45]), le16(msg[45:47]), le16(msg[47:49])
			if pc > 0 {
				copy(gotParams[pd:], msg[po:po+pc])
			}
			if dc > 0 {
				copy(gotData[dd:], msg[do:do+dc])
			}
		}
	}

	if !bytes.Equal(gotParams, params) {
		t.Errorf("reassembled request parameters do not match the original")
	}
	if !bytes.Equal(gotData, data) {
		t.Errorf("reassembled request data does not match the original")
	}
}

// TestTrans2SingleMessageNotFragmented confirms a small payload is sent as one
// SMB_COM_TRANSACTION2 message (no continuation messages).
func TestTrans2SingleMessageNotFragmented(t *testing.T) {
	cannedResp := makeTrans2ResponseMsg(0, nil, 0, 0, nil, 0)

	tr := &recordingTransport{response: cannedResp}
	c := &Client{
		Transport:  tr,
		Connection: &Connection{Server: &Server{MaxBufferSize: 4356}},
		Session:    &Session{SessionUID: 1, TreeID: 1},
	}

	if _, _, err := c.trans2(0x0001, []byte{0x16, 0x00, 0x00, 0x02}, nil); err != nil {
		t.Fatalf("trans2 error: %v", err)
	}
	if len(tr.sent) != 1 {
		t.Fatalf("expected exactly one message, got %d", len(tr.sent))
	}
	if tr.sent[0][4] != byte(codes.SMB_COM_TRANSACTION2) {
		t.Errorf("command = 0x%02x, want TRANSACTION2", tr.sent[0][4])
	}
}

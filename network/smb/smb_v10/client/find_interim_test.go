package client

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// scriptedTransport replays queued responses and records the order in which the
// client sent and received, so a test can tell whether the client waited for a
// response before sending the next request.
type scriptedTransport struct {
	replies [][]byte
	events  []string
}

func (s *scriptedTransport) Connect(net.IP, int) error { return nil }
func (s *scriptedTransport) Close() error              { return nil }
func (s *scriptedTransport) IsConnected() bool         { return true }
func (s *scriptedTransport) SetTimeout(time.Duration)  {}
func (s *scriptedTransport) Send(data []byte) (int, error) {
	s.events = append(s.events, fmt.Sprintf("send 0x%02X", data[4]))
	return len(data), nil
}
func (s *scriptedTransport) Receive() ([]byte, error) {
	if len(s.replies) == 0 {
		s.events = append(s.events, "receive with nothing queued")
		return nil, fmt.Errorf("no reply queued")
	}
	reply := s.replies[0]
	s.replies = s.replies[1:]
	s.events = append(s.events, fmt.Sprintf("receive 0x%08X", binary.LittleEndian.Uint32(reply[5:9])))
	return reply, nil
}

// interimTrans2Response builds the interim response a server sends to a primary
// TRANSACTION2 that does not carry the whole transaction: the header with the
// given status, and WordCount and ByteCount of zero.
func interimTrans2Response(t *testing.T, status uint32) []byte {
	t.Helper()

	final := finalTrans2Response(t)
	interim := append([]byte(nil), final[:32]...)
	binary.LittleEndian.PutUint32(interim[5:9], status)
	return append(interim, 0x00, 0x00, 0x00)
}

// finalTrans2Response builds a TRANSACTION2 response carrying the data bytes
// DE AD.
func finalTrans2Response(t *testing.T) []byte {
	t.Helper()

	response := commands.NewTransaction2Response()
	response.TotalDataCount = 2
	response.DataCount = 2
	response.DataOffset = 55
	response.Trans2_Data = []types.UCHAR{0xDE, 0xAD}
	msg := message.NewMessage()
	msg.Header.SetFlags(flags.FLAGS_REPLY)
	msg.AddCommand(response)
	raw, err := msg.Marshal()
	if err != nil {
		t.Fatalf("marshal the final response: %v", err)
	}
	return raw
}

// TestFragmentedTrans2WaitsForTheInterimResponse asserts the client reads the
// server's interim response before sending any secondary request, and takes the
// response after the secondaries as the answer.
//
// [MS-CIFS] section 3.2.4.1.5: "If an interim response is sent, and no error is
// returned in the interim response, then a "secondary request" MUST be used to
// continue a transaction started with a primary request". A client that sent the
// secondaries first read the interim response as the answer, and returned its
// empty blocks as the transaction's result without reporting anything wrong.
func TestFragmentedTrans2WaitsForTheInterimResponse(t *testing.T) {
	transport := &scriptedTransport{replies: [][]byte{interimTrans2Response(t, 0), finalTrans2Response(t)}}
	c := &Client{
		Transport:  transport,
		Connection: &Connection{Server: &Server{MaxBufferSize: 192}},
		Session:    &Session{},
	}

	_, data, err := c.trans2(0x0001, make([]byte, 100), make([]byte, 300))
	if err != nil {
		t.Fatalf("trans2: %v (events: %v)", err, transport.events)
	}
	if string(data) != "\xDE\xAD" {
		t.Errorf("trans2 returned data % x, want de ad (events: %v)", data, transport.events)
	}

	if len(transport.events) < 4 ||
		transport.events[0] != "send 0x32" ||
		transport.events[1] != "receive 0x00000000" ||
		transport.events[2] != "send 0x33" ||
		transport.events[len(transport.events)-1] != "receive 0x00000000" {
		t.Fatalf("the exchange ran %v, want the primary, the interim response, the secondaries, then the answer",
			transport.events)
	}
	for _, event := range transport.events[2 : len(transport.events)-1] {
		if event != "send 0x33" {
			t.Errorf("between the interim response and the answer the client did %q, want only secondaries", event)
		}
	}
}

// TestFragmentedTrans2StopsOnAnInterimError asserts an interim response that
// carries an error ends the transaction without sending the secondaries.
//
// [MS-CIFS] section 3.2.5.1.4: "If the interim response indicates an error, then
// the transaction is canceled. The client MUST NOT send any secondary
// transaction request messages."
func TestFragmentedTrans2StopsOnAnInterimError(t *testing.T) {
	const insufficientResources = 0xC0000205
	transport := &scriptedTransport{replies: [][]byte{interimTrans2Response(t, insufficientResources)}}
	c := &Client{
		Transport:  transport,
		Connection: &Connection{Server: &Server{MaxBufferSize: 192}},
		Session:    &Session{},
	}

	_, _, err := c.trans2(0x0001, make([]byte, 100), make([]byte, 300))
	if err == nil || !strings.Contains(err.Error(), "c0000205") {
		t.Errorf("trans2 error = %v, want one reporting 0xc0000205", err)
	}
	for _, event := range transport.events {
		if event == "send 0x33" {
			t.Fatalf("a secondary was sent after an interim error (events: %v)", transport.events)
		}
	}
}

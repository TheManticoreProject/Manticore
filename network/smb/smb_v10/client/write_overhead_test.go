package client_test

import (
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// largestSendTransport answers every request with the same response and records the
// size of the largest request sent.
type largestSendTransport struct {
	largest  int
	response []byte
}

func (m *largestSendTransport) Connect(ipaddr net.IP, port int) error { return nil }
func (m *largestSendTransport) Close() error                          { return nil }
func (m *largestSendTransport) Send(data []byte) (int, error) {
	if len(data) > m.largest {
		m.largest = len(data)
	}
	return len(data), nil
}
func (m *largestSendTransport) Receive() ([]byte, error) { return m.response, nil }
func (m *largestSendTransport) IsConnected() bool        { return true }
func (m *largestSendTransport) SetTimeout(time.Duration) {}

// At offsets of 4 GiB and above WriteAndx carries OffsetHigh (14 parameter words
// instead of 12), so the chunk must shrink to keep the request within MaxBufferSize.
func TestWriteFileChunkFitsMaxBufferAbove4GiB(t *testing.T) {
	const maxBuf = 16644 // Windows default MaxBufferSize
	for _, offset := range []uint64{0, 0x1_0000_0000} {
		resp := commands.NewWriteAndxResponse()
		resp.Count = types.USHORT(0x1000)
		tr := &largestSendTransport{response: marshalResponse(t, resp)}
		c := &client.Client{
			Transport:  tr,
			Connection: &client.Connection{Server: &client.Server{MaxBufferSize: maxBuf}},
			Session:    &client.Session{SessionUID: 1, TreeID: 1},
		}
		if _, err := c.WriteFile(client.FID(1), offset, make([]byte, 2*maxBuf)); err != nil {
			t.Fatalf("offset %#x: WriteFile: %v", offset, err)
		}
		if tr.largest > maxBuf {
			t.Errorf("offset %#x: largest WriteAndx request is %d bytes, exceeds MaxBufferSize %d", offset, tr.largest, maxBuf)
		}
	}
}

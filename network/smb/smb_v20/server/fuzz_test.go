package server

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

// discardTransport accepts every write; the fuzzer only cares about decoding.
type discardTransport struct{ transport.Transport }

func (discardTransport) Send(b []byte) (int, error) { return len(b), nil }
func (discardTransport) Close() error               { return nil }

// FuzzHandleFrame feeds arbitrary frames to a negotiated connection, which must
// answer or reject them without panicking.
func FuzzHandleFrame(f *testing.F) {
	negotiate := commands.NewNegotiateRequest()
	negotiate.AddDialect(dialects.SMB2_DIALECT_3_1_1)
	negotiate.Contexts = []*commands.NegotiateContext{commands.NewPreauthIntegrityContext(make([]byte, 32))}
	m := message.NewMessage()
	m.SetCommand(negotiate)
	seed, _ := m.Marshal()
	f.Add(seed)
	f.Add([]byte{0xFF, 'S', 'M', 'B', 0x72})
	f.Add([]byte{0xFD, 'S', 'M', 'B'})

	f.Fuzz(func(t *testing.T, frame []byte) {
		srv, err := NewServer(Config{})
		if err != nil {
			t.Fatal(err)
		}
		c := newConnection(srv, discardTransport{}, nil)
		c.handleFrame(frame)
		c.dialect, c.credits = dialects.SMB2_DIALECT_3_1_1, creditWindow{granted: 1 << 20, available: maxCredits}
		c.handleFrame(frame)
	})
}

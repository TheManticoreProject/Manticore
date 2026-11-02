package client

import (
	"encoding/binary"
	"net"
	"testing"
	"time"

	smb1 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	smb1message "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	smb1commands "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	smb1flags "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	smb2 "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/client"
	smb2message "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	smb2commands "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	smb2flags "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// replayTransport records each request and answers it with the next canned
// response, stamping the request's SMB2 MessageId onto it.
type replayTransport struct {
	sent      [][]byte
	responses [][]byte
	smb2      bool
}

func (r *replayTransport) Connect(net.IP, int) error { return nil }
func (r *replayTransport) Close() error              { return nil }
func (r *replayTransport) IsConnected() bool         { return true }
func (r *replayTransport) SetTimeout(time.Duration)  {}
func (r *replayTransport) Send(data []byte) (int, error) {
	r.sent = append(r.sent, append([]byte(nil), data...))
	return len(data), nil
}
func (r *replayTransport) Receive() ([]byte, error) {
	resp := r.responses[0]
	r.responses = r.responses[1:]
	if r.smb2 && len(r.sent) > 0 {
		copy(resp[24:32], r.sent[len(r.sent)-1][24:32])
	}
	return resp, nil
}

// TestOpenFilePassesFileAttributes drives OpenFile through both adapters with
// OpenOptions.FileAttributes set and decodes the create request each engine
// sent: the adapters used to drop the field, so every file was created
// FILE_ATTRIBUTE_NORMAL.
func TestOpenFilePassesFileAttributes(t *testing.T) {
	opts := OpenOptions{
		DesiredAccess:     fileflags.GENERIC_WRITE,
		CreateDisposition: fileflags.FILE_OVERWRITE_IF,
		CreateOptions:     fileflags.FILE_NON_DIRECTORY_FILE,
		FileAttributes:    fileflags.FILE_ATTRIBUTE_HIDDEN,
	}

	t.Run("SMB2", func(t *testing.T) {
		resp := smb2commands.NewCreateResponse()
		m := smb2message.NewMessage()
		m.Header.AddFlags(smb2flags.SMB2_FLAGS_SERVER_TO_REDIR)
		m.Header.TreeId, m.Header.SessionId = 5, 0x99
		m.SetCommand(resp)
		wire, err := m.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		tr := &replayTransport{responses: [][]byte{wire}, smb2: true}
		engine := smb2.NewFromTransport(tr, net.ParseIP("127.0.0.1"), 445)
		engine.Session = &smb2.Session{Client: engine, SessionId: 0x99, TreeId: 5}
		engine.Connection.SessionTable[0x99] = engine.Session

		if _, err := newSMB2Backend(engine).OpenFile("f.txt", opts); err != nil {
			t.Fatalf("OpenFile: %v", err)
		}
		req := smb2commands.NewCreateRequest()
		if _, err := req.Unmarshal(tr.sent[0][64:]); err != nil {
			t.Fatalf("decode CREATE: %v", err)
		}
		if req.FileAttributes != fileflags.FILE_ATTRIBUTE_HIDDEN {
			t.Errorf("CREATE FileAttributes = 0x%08x, want FILE_ATTRIBUTE_HIDDEN", req.FileAttributes)
		}
	})

	t.Run("SMB1", func(t *testing.T) {
		resp := smb1commands.NewNtCreateAndxResponse()
		m := smb1message.NewMessage()
		m.AddCommand(resp)
		m.Header.SetFlags(smb1flags.FLAGS_REPLY)
		wire, err := m.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		tr := &replayTransport{responses: [][]byte{wire}}
		engine := smb1.NewFromTransport(tr, net.ParseIP("127.0.0.1"), 445)
		engine.Session = &smb1.Session{SessionUID: 1, TreeID: 1}

		if _, err := newSMB1Backend(engine).OpenFile("f.txt", opts); err != nil {
			t.Fatalf("OpenFile: %v", err)
		}
		reqMsg := smb1message.NewMessage()
		if err := reqMsg.Unmarshal(tr.sent[0]); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		req, ok := reqMsg.Command.(*smb1commands.NtCreateAndxRequest)
		if !ok {
			t.Fatalf("sent %T, want *NtCreateAndxRequest", reqMsg.Command)
		}
		if got := binary.LittleEndian.Uint32(binary.LittleEndian.AppendUint32(nil, uint32(req.ExtFileAttributes))); got != fileflags.FILE_ATTRIBUTE_HIDDEN {
			t.Errorf("NT_CREATE_ANDX ExtFileAttributes = 0x%08x, want FILE_ATTRIBUTE_HIDDEN", got)
		}
	})
}

package server

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
)

// sinkTransport is a Transport that discards what it is given and never yields a
// frame. It stands in for a client while a fuzz target drives the frame handler
// directly, without a socket in the way.
type sinkTransport struct {
	sent [][]byte
}

func (t *sinkTransport) Connect(net.IP, int) error { return nil }
func (t *sinkTransport) Close() error              { return nil }
func (t *sinkTransport) Send(data []byte) (int, error) {
	t.sent = append(t.sent, append([]byte(nil), data...))
	return len(data), nil
}
func (t *sinkTransport) Receive() ([]byte, error) { return nil, io.EOF }
func (t *sinkTransport) IsConnected() bool        { return true }
func (t *sinkTransport) SetTimeout(time.Duration) {}

// FuzzServerFrame drives arbitrary bytes through the frame handler, which is the
// first thing a listening service exposes to the network. The handler is allowed
// to answer, to drop the connection, or to do neither; it is not allowed to
// panic, and it is not allowed to read or allocate outside the frame it was
// given.
func FuzzServerFrame(f *testing.F) {
	// Seed with the shapes the handler distinguishes, so the fuzzer starts from
	// each branch rather than having to rediscover the SMB header.
	f.Add([]byte{})
	f.Add([]byte{0xFF, 'S', 'M', 'B'})
	f.Add(append([]byte{0xFE, 'S', 'M', 'B'}, make([]byte, 31)...))

	// A well-formed echo request.
	if marshalled, err := echoRequest("fuzz seed").Marshal(); err == nil {
		f.Add(marshalled)
	}
	// A recognized command with no handler.
	unimplemented := newRequest(codes.SMB_COM_SEEK)
	unimplemented.AddCommand(commands.NewSeekRequest())
	if marshalled, err := unimplemented.Marshal(); err == nil {
		f.Add(marshalled)
	}
	// A header claiming to be a response.
	f.Add(rawFrame(codes.SMB_COM_ECHO, flags.Flags(flags.FLAGS_REPLY), flags2.Flags2(0), []byte{0x00, 0x00, 0x00}))
	// A body whose WordCount overruns the frame.
	f.Add(rawFrame(codes.SMB_COM_ECHO, flags.Flags(0), flags2.Flags2(0), []byte{0xFF, 0x00, 0x00}))
	// An unrecognized command code.
	if unknown, ok := firstUnknownCommandCode(); ok {
		f.Add(rawFrame(unknown, flags.Flags(0), flags2.Flags2(0), []byte{0x00, 0x00, 0x00}))
	}
	// An AndX chain whose offset points back at itself, which the message layer
	// has to reject rather than follow forever.
	f.Add(rawFrame(codes.SMB_COM_SESSION_SETUP_ANDX, flags.Flags(0), flags2.Flags2(0),
		[]byte{0x0D, byte(codes.SMB_COM_TREE_CONNECT_ANDX), 0x00, 0x20, 0x00}))

	// The session-setup layouts, which the authentication phases made reachable.
	// The extended-security form is chosen by WordCount 12 and the password form
	// by WordCount 10, and each carries client-controlled lengths that the
	// decoder has to bound.
	f.Add(rawFrame(codes.SMB_COM_SESSION_SETUP_ANDX, flags.Flags(0),
		flags2.Flags2(flags2.FLAGS2_EXTENDED_SECURITY),
		append([]byte{0x0C}, make([]byte, 12*2+2)...)))
	f.Add(rawFrame(codes.SMB_COM_SESSION_SETUP_ANDX, flags.Flags(0), flags2.Flags2(0),
		append([]byte{0x0A}, make([]byte, 10*2+2)...)))

	// A logoff, which reaches the session table.
	f.Add(rawFrame(codes.SMB_COM_LOGOFF_ANDX, flags.Flags(0),
		flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES),
		[]byte{0x02, byte(codes.SMB_COM_NO_ANDX_COMMAND), 0x00, 0x00, 0x00, 0x00, 0x00}))

	// A frame announcing a signature on a connection that is not signing, and one
	// with the signature field filled in, since the verification path reads that
	// field before anything else looks at the body.
	f.Add(rawFrame(codes.SMB_COM_ECHO, flags.Flags(0),
		flags2.Flags2(flags2.FLAGS2_SECURITY_SIGNATURE),
		[]byte{0x01, 0x01, 0x00, 0x00, 0x00}))

	// A TRANSACTION2 whose declared totals exceed what it carries, which is what
	// the reassembly arithmetic has to refuse rather than allocate for, and a
	// secondary message with no transaction in progress.
	f.Add(rawFrame(codes.SMB_COM_TRANSACTION2, flags.Flags(0),
		flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES),
		append([]byte{0x0F}, make([]byte, 15*2+2)...)))
	f.Add(rawFrame(codes.SMB_COM_TRANSACTION2_SECONDARY, flags.Flags(0),
		flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES),
		append([]byte{0x09}, make([]byte, 9*2+2)...)))

	// The file commands, each carrying a client-controlled length or offset.
	for _, command := range []codes.CommandCode{
		codes.SMB_COM_NT_CREATE_ANDX, codes.SMB_COM_READ_ANDX, codes.SMB_COM_WRITE_ANDX,
		codes.SMB_COM_TREE_CONNECT_ANDX, codes.SMB_COM_DELETE, codes.SMB_COM_RENAME,
	} {
		f.Add(rawFrame(command, flags.Flags(0),
			flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES),
			append([]byte{0x0C}, make([]byte, 12*2+2)...)))
	}

	// The NT_TRANSACT family, whose counts are 32-bit: a declared total is an
	// allocation, so the bound on it is reachable from the first frame a client
	// sends. WordCount 19 selects the primary form, 18 the secondary.
	f.Add(rawFrame(codes.SMB_COM_NT_TRANSACT, flags.Flags(0),
		flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES),
		append([]byte{0x13}, make([]byte, 19*2+2)...)))
	// The same, with every count field set to 0xFFFFFFFF.
	f.Add(rawFrame(codes.SMB_COM_NT_TRANSACT, flags.Flags(0),
		flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES),
		append([]byte{0x13}, bytes.Repeat([]byte{0xFF}, 19*2+2)...)))
	f.Add(rawFrame(codes.SMB_COM_NT_TRANSACT_SECONDARY, flags.Flags(0),
		flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES),
		append([]byte{0x12}, make([]byte, 18*2+2)...)))
	f.Add(rawFrame(codes.SMB_COM_NT_CANCEL, flags.Flags(0),
		flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES), []byte{0x00, 0x00, 0x00}))

	// The TRANSACTION family, whose data block begins with a name the decoder has
	// to find the end of. A block with no terminator at all is the case that a
	// scan without a bound would run off.
	f.Add(rawFrame(codes.SMB_COM_TRANSACTION, flags.Flags(0),
		flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES),
		append([]byte{0x0E}, make([]byte, 14*2+2)...)))
	transactionWithName := append([]byte{0x10}, make([]byte, 16*2)...)
	// SetupCount at word 14 (byte offset 1+28), then two setup words, then a
	// ByteCount naming a block of unterminated name bytes.
	transactionWithName[1+26] = 0x02
	transactionWithName = append(transactionWithName, 0x0C, 0x00)
	transactionWithName = append(transactionWithName, []byte(`\PIPE\srvsvc`)...)
	f.Add(rawFrame(codes.SMB_COM_TRANSACTION, flags.Flags(0),
		flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES), transactionWithName))
	// The same declared as Unicode, where the terminator is two bytes wide and an
	// odd-length block cannot hold one.
	f.Add(rawFrame(codes.SMB_COM_TRANSACTION, flags.Flags(0),
		flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES|flags2.FLAGS2_UNICODE), transactionWithName))
	f.Add(rawFrame(codes.SMB_COM_TRANSACTION_SECONDARY, flags.Flags(0),
		flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES),
		append([]byte{0x08}, make([]byte, 8*2+2)...)))

	f.Fuzz(func(t *testing.T, raw []byte) {
		srv, err := NewServer(Config{})
		if err != nil {
			t.Fatalf("NewServer() error = %v", err)
		}
		conn := newConnection(srv, &sinkTransport{}, &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1})

		// The contract is only that this returns: an error means the connection
		// would be dropped, nil means it would be kept. Either is fine for
		// arbitrary input; panicking or hanging is not.
		_ = conn.handleFrame(raw)
	})
}

// TestHandleFrameNeverPanicsOnTruncation walks every prefix of a valid request
// through the frame handler. A truncation boundary is where a length check that
// was written one byte off shows up, and it is reachable by any client.
func TestHandleFrameNeverPanicsOnTruncation(t *testing.T) {
	marshalled, err := echoRequest("truncate me").Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the request: %v", err)
	}

	srv, err := NewServer(Config{})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	for length := 0; length <= len(marshalled); length++ {
		length := length
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("handleFrame panicked on a %d-byte prefix: %v", length, r)
				}
			}()
			conn := newConnection(srv, &sinkTransport{}, &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1})
			_ = conn.handleFrame(marshalled[:length])
		}()
	}
}

// TestHandleFrameNeverPanicsOnCorruption flips each byte of a valid request in
// turn and pushes the result through the frame handler, covering the fields the
// truncation walk leaves intact.
func TestHandleFrameNeverPanicsOnCorruption(t *testing.T) {
	marshalled, err := echoRequest("corrupt me").Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the request: %v", err)
	}

	srv, err := NewServer(Config{})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	for index := range marshalled {
		for _, replacement := range []byte{0x00, 0x01, 0x7F, 0xFF} {
			index, replacement := index, replacement
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("handleFrame panicked with byte %d set to 0x%02X: %v", index, replacement, r)
					}
				}()
				corrupted := append([]byte(nil), marshalled...)
				corrupted[index] = replacement
				conn := newConnection(srv, &sinkTransport{}, &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1})
				_ = conn.handleFrame(corrupted)
			}()
		}
	}
}

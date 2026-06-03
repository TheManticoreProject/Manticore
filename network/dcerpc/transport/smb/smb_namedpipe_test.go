package smb

import (
	"bytes"
	"errors"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
)

// fakePipeConn is an in-memory pipeConn for exercising the transport without a live
// SMB server. It records the arguments it was called with and returns canned results.
type fakePipeConn struct {
	openFID client.FID
	openErr error

	writeErr error
	writeN   int // if zero, defaults to len(data)

	readData []byte
	readErr  error

	closeErr error

	// Recorded calls.
	openedPath  string
	openCount   int
	written     []byte
	writeOffset uint64
	readOffset  uint64
	readMaxLen  uint32
	closeCount  int
}

func (f *fakePipeConn) OpenFile(path string, desiredAccess, shareAccess, createDisp, createOptions uint32) (client.FID, error) {
	f.openCount++
	f.openedPath = path
	if f.openErr != nil {
		return 0, f.openErr
	}
	return f.openFID, nil
}

func (f *fakePipeConn) WriteFile(fid client.FID, offset uint64, data []byte) (int, error) {
	f.written = append([]byte(nil), data...)
	f.writeOffset = offset
	if f.writeErr != nil {
		return 0, f.writeErr
	}
	if f.writeN != 0 {
		return f.writeN, nil
	}
	return len(data), nil
}

func (f *fakePipeConn) ReadFile(fid client.FID, offset uint64, maxLen uint32) ([]byte, error) {
	f.readOffset = offset
	f.readMaxLen = maxLen
	if f.readErr != nil {
		return nil, f.readErr
	}
	return f.readData, nil
}

func (f *fakePipeConn) CloseFile(fid client.FID) error {
	f.closeCount++
	return f.closeErr
}

func TestNew_NormalizesPipeName(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want string
	}{
		{"PIPE\\lsarpc", "\\PIPE\\lsarpc"},
		{"\\PIPE\\lsarpc", "\\PIPE\\lsarpc"},
		{"", "\\"},
	} {
		p := New(nil, tc.in)
		if p.pipeName != tc.want {
			t.Errorf("New(%q): pipeName = %q, want %q", tc.in, p.pipeName, tc.want)
		}
	}
}

func TestNew_DefaultFragmentSizes(t *testing.T) {
	p := New(nil, `\PIPE\lsarpc`)
	if p.MaxXmitFrag() != DefaultMaxXmitFrag {
		t.Errorf("MaxXmitFrag() = %d, want %d", p.MaxXmitFrag(), DefaultMaxXmitFrag)
	}
	if p.MaxRecvFrag() != DefaultMaxRecvFrag {
		t.Errorf("MaxRecvFrag() = %d, want %d", p.MaxRecvFrag(), DefaultMaxRecvFrag)
	}

	p.SetMaxFrag(0xFFFF, 0x1000)
	if p.MaxXmitFrag() != 0xFFFF || p.MaxRecvFrag() != 0x1000 {
		t.Errorf("after SetMaxFrag: xmit=%d recv=%d, want 65535/4096", p.MaxXmitFrag(), p.MaxRecvFrag())
	}
}

func TestConnect_OpensPipeWithExpectedArgs(t *testing.T) {
	fake := &fakePipeConn{openFID: 0x4242}
	p := newWithConn(fake, `\PIPE\lsarpc`)

	if err := p.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if fake.openCount != 1 {
		t.Fatalf("OpenFile called %d times, want 1", fake.openCount)
	}
	if fake.openedPath != `\PIPE\lsarpc` {
		t.Errorf("OpenFile path = %q, want %q", fake.openedPath, `\PIPE\lsarpc`)
	}
	if p.fid != 0x4242 || !p.opened {
		t.Errorf("after Connect: fid=%#x opened=%v, want 0x4242/true", p.fid, p.opened)
	}
}

func TestConnect_Idempotent(t *testing.T) {
	fake := &fakePipeConn{openFID: 1}
	p := newWithConn(fake, `\PIPE\lsarpc`)

	if err := p.Connect(); err != nil {
		t.Fatalf("first Connect() error = %v", err)
	}
	if err := p.Connect(); err != nil {
		t.Fatalf("second Connect() error = %v", err)
	}
	if fake.openCount != 1 {
		t.Errorf("OpenFile called %d times, want 1 (Connect must be idempotent)", fake.openCount)
	}
}

func TestConnect_PropagatesOpenError(t *testing.T) {
	wantErr := errors.New("access denied")
	fake := &fakePipeConn{openErr: wantErr}
	p := newWithConn(fake, `\PIPE\lsarpc`)

	err := p.Connect()
	if err == nil {
		t.Fatal("Connect() error = nil, want non-nil")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("Connect() error = %v, want wrapped %v", err, wantErr)
	}
	if p.opened {
		t.Error("transport marked opened after a failed Connect")
	}
}

func TestSendReceive_WritesThenReads(t *testing.T) {
	resp := []byte{0x05, 0x00, 0x0c, 0x03} // looks like a bind_ack header prefix
	fake := &fakePipeConn{openFID: 7, readData: resp}
	p := newWithConn(fake, `\PIPE\lsarpc`)
	if err := p.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	req := []byte{0x05, 0x00, 0x0b, 0x03, 0xde, 0xad}
	got, err := p.SendReceive(req)
	if err != nil {
		t.Fatalf("SendReceive() error = %v", err)
	}

	if !bytes.Equal(fake.written, req) {
		t.Errorf("written = %x, want %x", fake.written, req)
	}
	if fake.writeOffset != 0 || fake.readOffset != 0 {
		t.Errorf("offsets: write=%d read=%d, want 0/0 (pipe offset is meaningless)", fake.writeOffset, fake.readOffset)
	}
	if fake.readMaxLen != uint32(DefaultMaxRecvFrag) {
		t.Errorf("read maxLen = %d, want %d", fake.readMaxLen, DefaultMaxRecvFrag)
	}
	if !bytes.Equal(got, resp) {
		t.Errorf("SendReceive() = %x, want %x", got, resp)
	}
}

func TestSendReceive_RequiresConnect(t *testing.T) {
	p := newWithConn(&fakePipeConn{}, `\PIPE\lsarpc`)
	if _, err := p.SendReceive([]byte{0x01}); err == nil {
		t.Fatal("SendReceive() before Connect: error = nil, want non-nil")
	}
}

func TestSendReceive_RejectsEmptyPDU(t *testing.T) {
	fake := &fakePipeConn{openFID: 1}
	p := newWithConn(fake, `\PIPE\lsarpc`)
	if err := p.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if _, err := p.SendReceive(nil); err == nil {
		t.Fatal("SendReceive(nil): error = nil, want non-nil")
	}
	if fake.written != nil {
		t.Error("empty PDU should not have been written")
	}
}

func TestSendReceive_ShortWriteIsError(t *testing.T) {
	fake := &fakePipeConn{openFID: 1, writeN: 2, readData: []byte{0xff}}
	p := newWithConn(fake, `\PIPE\lsarpc`)
	if err := p.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if _, err := p.SendReceive([]byte{0x01, 0x02, 0x03, 0x04}); err == nil {
		t.Fatal("SendReceive() with short write: error = nil, want non-nil")
	}
}

func TestSendReceive_PropagatesReadError(t *testing.T) {
	wantErr := errors.New("pipe broken")
	fake := &fakePipeConn{openFID: 1, readErr: wantErr}
	p := newWithConn(fake, `\PIPE\lsarpc`)
	if err := p.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	_, err := p.SendReceive([]byte{0x01})
	if !errors.Is(err, wantErr) {
		t.Errorf("SendReceive() error = %v, want wrapped %v", err, wantErr)
	}
}

func TestSendReceive_EmptyResponseIsError(t *testing.T) {
	fake := &fakePipeConn{openFID: 1, readData: []byte{}}
	p := newWithConn(fake, `\PIPE\lsarpc`)
	if err := p.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if _, err := p.SendReceive([]byte{0x01}); err == nil {
		t.Fatal("SendReceive() with empty response: error = nil, want non-nil")
	}
}

func TestClose_ClosesPipeAndIsIdempotent(t *testing.T) {
	fake := &fakePipeConn{openFID: 1}
	p := newWithConn(fake, `\PIPE\lsarpc`)
	if err := p.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if err := p.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if fake.closeCount != 1 {
		t.Errorf("CloseFile called %d times, want 1", fake.closeCount)
	}
	if p.opened {
		t.Error("transport still marked opened after Close")
	}

	// Second Close is a no-op.
	if err := p.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if fake.closeCount != 1 {
		t.Errorf("CloseFile called %d times after second Close, want 1", fake.closeCount)
	}
}

func TestClose_BeforeConnectIsNoOp(t *testing.T) {
	fake := &fakePipeConn{}
	p := newWithConn(fake, `\PIPE\lsarpc`)
	if err := p.Close(); err != nil {
		t.Fatalf("Close() before Connect error = %v", err)
	}
	if fake.closeCount != 0 {
		t.Errorf("CloseFile called %d times, want 0", fake.closeCount)
	}
}

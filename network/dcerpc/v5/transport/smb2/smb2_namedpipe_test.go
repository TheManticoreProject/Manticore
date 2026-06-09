package smb2

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// fakePipe records the operations the transport performs.
type fakePipe struct {
	created       bool
	writes        [][]byte
	transceiveIn  [][]byte
	reads         int
	closed        bool
	transceiveOut []byte
	readOut       []byte
}

func (f *fakePipe) CreateFile(string, uint32, uint32, uint32, uint32) (types.SMB2_FILEID, error) {
	f.created = true
	return types.SMB2_FILEID{}, nil
}
func (f *fakePipe) TransactNamedPipe(_ types.SMB2_FILEID, input []byte, _ uint32) ([]byte, error) {
	f.transceiveIn = append(f.transceiveIn, append([]byte(nil), input...))
	return f.transceiveOut, nil
}
func (f *fakePipe) WriteFile(_ types.SMB2_FILEID, _ uint64, data []byte) (uint32, error) {
	f.writes = append(f.writes, append([]byte(nil), data...))
	return uint32(len(data)), nil
}
func (f *fakePipe) ReadFile(types.SMB2_FILEID, uint64, uint32) ([]byte, error) {
	f.reads++
	return f.readOut, nil
}
func (f *fakePipe) CloseFile(types.SMB2_FILEID) error { f.closed = true; return nil }

func TestSingleFragment_UsesTransceive(t *testing.T) {
	f := &fakePipe{transceiveOut: []byte("ack")}
	p := newWithConn(f, "samr")
	if err := p.Connect(); err != nil || !f.created {
		t.Fatalf("Connect: err=%v created=%v", err, f.created)
	}
	if err := p.Send([]byte("bind")); err != nil {
		t.Fatal(err)
	}
	out, err := p.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if len(f.writes) != 0 {
		t.Errorf("expected no plain writes for a single fragment, got %d", len(f.writes))
	}
	if len(f.transceiveIn) != 1 || !bytes.Equal(f.transceiveIn[0], []byte("bind")) {
		t.Errorf("expected one transceive of \"bind\", got %v", f.transceiveIn)
	}
	if !bytes.Equal(out, []byte("ack")) {
		t.Errorf("Recv = %q, want \"ack\"", out)
	}
}

func TestMultiFragment_FlushesPriorViaWrite(t *testing.T) {
	f := &fakePipe{transceiveOut: []byte("resp")}
	p := newWithConn(f, "samr")
	p.Connect()
	p.Send([]byte("frag1"))
	p.Send([]byte("frag2")) // flushes frag1 via a plain write
	out, err := p.Recv()    // transceives frag2
	if err != nil {
		t.Fatal(err)
	}
	if len(f.writes) != 1 || !bytes.Equal(f.writes[0], []byte("frag1")) {
		t.Errorf("expected frag1 written via WriteFile, got %v", f.writes)
	}
	if len(f.transceiveIn) != 1 || !bytes.Equal(f.transceiveIn[0], []byte("frag2")) {
		t.Errorf("expected frag2 transceived, got %v", f.transceiveIn)
	}
	if !bytes.Equal(out, []byte("resp")) {
		t.Errorf("Recv = %q, want \"resp\"", out)
	}
}

func TestContinuationRead(t *testing.T) {
	f := &fakePipe{transceiveOut: []byte("first"), readOut: []byte("more")}
	p := newWithConn(f, "samr")
	p.Connect()
	p.Send([]byte("req"))
	if _, err := p.Recv(); err != nil { // transceive
		t.Fatal(err)
	}
	out, err := p.Recv() // no pending -> ReadFile
	if err != nil {
		t.Fatal(err)
	}
	if f.reads != 1 || !bytes.Equal(out, []byte("more")) {
		t.Errorf("continuation read: reads=%d out=%q", f.reads, out)
	}
}

func TestPipeNameNormalization(t *testing.T) {
	for _, in := range []string{`\PIPE\lsarpc`, `\lsarpc`, `PIPE\lsarpc`, "lsarpc"} {
		if p := newWithConn(&fakePipe{}, in); p.pipeName != "lsarpc" {
			t.Errorf("pipe name %q normalized to %q, want \"lsarpc\"", in, p.pipeName)
		}
	}
}

func TestSendBeforeConnect(t *testing.T) {
	p := newWithConn(&fakePipe{}, "samr")
	if err := p.Send([]byte("x")); err == nil {
		t.Error("expected error sending before Connect")
	}
}

package server

import (
	"bytes"
	"strings"
	"testing"
)

// TestNamedPipeWriteThenRead covers the exchange a client that does not
// transact performs: write the message with WRITE_ANDX, then collect the answer
// with READ_ANDX. impacket's DCE/RPC over SMB works this way, and the server
// refused the write with STATUS_NOT_SUPPORTED.
func TestNamedPipeWriteThenRead(t *testing.T) {
	pipes := newEchoPipe("srvsvc")
	_, client := pipeServer(t, pipes)

	fid, err := openPipeHandle(t, client, `\PIPE\srvsvc`)
	if err != nil {
		t.Fatalf("opening the pipe failed: %v", err)
	}
	payload := []byte("an rpc request would go here")
	written, err := client.WriteFile(fid, 0, payload)
	if err != nil || written != len(payload) {
		t.Fatalf("WriteFile() = %d, %v; want %d", written, err, len(payload))
	}
	answer, err := client.ReadFile(fid, 0, 4096)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if want := append([]byte("echo:"), payload...); !bytes.Equal(answer, want) {
		t.Fatalf("the pipe returned %q, want %q", answer, want)
	}

	// The next message is a new exchange on the same session.
	if _, err := client.WriteFile(fid, 0, []byte("second")); err != nil {
		t.Fatalf("second WriteFile() error = %v", err)
	}
	if answer, err := client.ReadFile(fid, 0, 4096); err != nil || string(answer) != "echo:second" {
		t.Fatalf("second ReadFile() = %q, %v", answer, err)
	}
	if got := pipes.sessionAt(0).transactCount(); got != 2 {
		t.Errorf("the session saw %d messages, want 2", got)
	}
}

// TestNamedPipeWriteWithUnreadAnswerIsRefused checks that a write while an
// answer is still unread is refused with STATUS_PIPE_BUSY, as a transaction is:
// otherwise a client that never reads would grow the buffered answers without
// limit.
func TestNamedPipeWriteWithUnreadAnswerIsRefused(t *testing.T) {
	pipes := newEchoPipe("srvsvc")
	_, client := pipeServer(t, pipes)

	fid, err := openPipeHandle(t, client, `\PIPE\srvsvc`)
	if err != nil {
		t.Fatalf("opening the pipe failed: %v", err)
	}
	if _, err := client.WriteFile(fid, 0, []byte("first")); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err = client.WriteFile(fid, 0, []byte("second"))
	if err == nil || !strings.Contains(err.Error(), "0xc00000ae") {
		t.Fatalf("a write over an unread answer: err = %v, want STATUS_PIPE_BUSY (0xc00000ae)", err)
	}
	if answer, err := client.ReadFile(fid, 0, 4096); err != nil || string(answer) != "echo:first" {
		t.Fatalf("ReadFile() = %q, %v; the first answer must survive the refused write", answer, err)
	}
}

// TestNamedPipeZeroLengthWriteIsANoOp checks an empty write succeeds without
// reaching the session, where empty input would mean "the rest of the answer".
func TestNamedPipeZeroLengthWriteIsANoOp(t *testing.T) {
	pipes := newEchoPipe("srvsvc")
	_, client := pipeServer(t, pipes)
	fid, err := openPipeHandle(t, client, `\PIPE\srvsvc`)
	if err != nil {
		t.Fatalf("opening the pipe failed: %v", err)
	}
	if written, err := client.WriteFile(fid, 0, nil); err != nil || written != 0 {
		t.Fatalf("an empty WriteFile() = %d, %v; want 0, nil", written, err)
	}
	if got := pipes.sessionAt(0).transactCount(); got != 0 {
		t.Errorf("an empty write reached the session %d times", got)
	}
}

// TestNamedPipeReadReturnsOneMessage checks a read never crosses from one
// message into the next: with an answer of three messages (two of 64 KiB and
// one of 500 bytes), every read lies within one message, and the reads together
// reassemble the answer.
func TestNamedPipeReadReturnsOneMessage(t *testing.T) {
	answer := countingAnswer(2*maxPipeAnswerSize + 500)
	_, client := pipeServer(t, &partedAnswerPipe{name: "srvsvc", answer: answer})
	fid, err := openPipeHandle(t, client, `\PIPE\srvsvc`)
	if err != nil {
		t.Fatalf("opening the pipe failed: %v", err)
	}
	if _, err := client.WriteFile(fid, 0, []byte("call")); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	var collected []byte
	for reads := 0; ; reads++ {
		part, err := client.ReadFile(fid, 0, uint32(len(answer)))
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}
		if len(part) == 0 {
			break
		}
		start, end := len(collected), len(collected)+len(part)-1
		if start/maxPipeAnswerSize != end/maxPipeAnswerSize {
			t.Fatalf("read %d spans bytes %d-%d, crossing a message boundary", reads, start, end)
		}
		collected = append(collected, part...)
	}
	if !bytes.Equal(collected, answer) {
		t.Fatalf("the answer reassembled to %d bytes, want %d", len(collected), len(answer))
	}
}

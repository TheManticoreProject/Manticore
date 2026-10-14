package hostserver_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

func newPipeConn(t *testing.T, opts hostserver.Options) (*hostserver.Conn, net.Conn) {
	t.Helper()
	client, server := net.Pipe()
	c, err := hostserver.NewConn(client, opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close(); server.Close() })
	return c, server
}

func request(corr uint32) hostserver.Frame {
	return hostserver.Frame{
		Header:   hostserver.Header{ServerID: hostserver.ServerIDRemoteCommand, CorrelationID: corr, RequestID: 0x1003},
		Template: []byte{1, 2},
		Variable: []byte{0, 0, 0, 7, 0x11, 0x03, 0xC1},
	}
}

func replyBytes(t *testing.T, serverID uint16, corr uint32, replyID uint16) []byte {
	t.Helper()
	b, err := hostserver.Frame{
		Header:   hostserver.Header{ServerID: serverID, CorrelationID: corr, RequestID: replyID},
		Template: []byte{0, 0},
	}.Marshal(0)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

var expect1 = hostserver.Expect{ServerID: hostserver.ServerIDRemoteCommand, ReplyIDs: []uint16{0x8003, 0x9003}, CorrelationID: 1}

// serveOne reads one request frame from server and writes reply.
func serveOne(t *testing.T, server net.Conn, reply []byte) <-chan []byte {
	got := make(chan []byte, 1)
	go func() {
		var hb [hostserver.HeaderLength]byte
		if _, err := io.ReadFull(server, hb[:]); err != nil {
			close(got)
			return
		}
		body := make([]byte, binary.BigEndian.Uint32(hb[:])-hostserver.HeaderLength)
		_, _ = io.ReadFull(server, body)
		got <- append(hb[:], body...)
		if reply != nil {
			_, _ = server.Write(reply)
		}
	}()
	return got
}

func wantPoisoned(t *testing.T, c *hostserver.Conn, err error) {
	t.Helper()
	if !errors.Is(err, hostserver.ErrPoisoned) {
		t.Fatalf("error = %v; want ErrPoisoned", err)
	}
	if !errors.Is(c.Err(), hostserver.ErrPoisoned) {
		t.Fatalf("Conn.Err() = %v after poisoning", c.Err())
	}
	if _, err := c.Send(context.Background(), request(9)); !errors.Is(err, hostserver.ErrPoisoned) {
		t.Fatalf("Send after poisoning = %v", err)
	}
}

func TestExchange(t *testing.T) {
	c, server := newPipeConn(t, hostserver.Options{})
	sent := serveOne(t, server, replyBytes(t, hostserver.ServerIDRemoteCommand, 1, 0x8003))
	reply, receipt, err := c.Exchange(context.Background(), request(1), expect1)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := request(1).Marshal(0)
	if got := <-sent; !bytes.Equal(got, want) || receipt.BytesAccepted != len(want) {
		t.Errorf("server received % X (receipt %d)", got, receipt.BytesAccepted)
	}
	if reply.Header.RequestID != 0x8003 || !bytes.Equal(reply.Template, []byte{0, 0}) || c.Err() != nil {
		t.Errorf("reply = %+v, Err %v", reply, c.Err())
	}
}

func TestExchangeRejectsUnexpectedReply(t *testing.T) {
	for name, reply := range map[string][]byte{
		"wrong server":      replyBytes(t, hostserver.ServerIDSignon, 1, 0x8003),
		"wrong correlation": replyBytes(t, hostserver.ServerIDRemoteCommand, 2, 0x8003),
		"wrong reply ID":    replyBytes(t, hostserver.ServerIDRemoteCommand, 1, 0x8001),
	} {
		t.Run(name, func(t *testing.T) {
			c, server := newPipeConn(t, hostserver.Options{})
			serveOne(t, server, reply)
			_, receipt, err := c.Exchange(context.Background(), request(1), expect1)
			if !errors.Is(err, hostserver.ErrUnexpectedReply) || receipt.BytesAccepted == 0 {
				t.Fatalf("error = %v, receipt %d", err, receipt.BytesAccepted)
			}
			wantPoisoned(t, c, err)
		})
	}
}

func TestSendCancelledBeforeStartLeavesConnUsable(t *testing.T) {
	c, server := newPipeConn(t, hostserver.Options{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	receipt, err := c.Send(ctx, request(1))
	if !errors.Is(err, hostserver.ErrSendNotStarted) || !errors.Is(err, context.Canceled) || receipt.BytesAccepted != 0 {
		t.Fatalf("Send = %d, %v", receipt.BytesAccepted, err)
	}
	if c.Err() != nil {
		t.Fatalf("Conn unusable after a not-started send: %v", c.Err())
	}
	serveOne(t, server, replyBytes(t, hostserver.ServerIDRemoteCommand, 1, 0x8003))
	if _, _, err := c.Exchange(context.Background(), request(1), expect1); err != nil {
		t.Fatalf("Exchange after not-started send: %v", err)
	}
}

func TestSendCancelledMidFramePoisons(t *testing.T) {
	c, server := newPipeConn(t, hostserver.Options{})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		buf := make([]byte, 5)
		_, _ = io.ReadFull(server, buf) // accept 5 bytes, then stall
		cancel()
	}()
	receipt, err := c.Send(ctx, request(1))
	if !errors.Is(err, context.Canceled) || receipt.BytesAccepted != 5 {
		t.Fatalf("Send = %d bytes, %v; want 5 bytes and cancellation", receipt.BytesAccepted, err)
	}
	wantPoisoned(t, c, err)
}

func TestReceiveCancelledMidHeaderAndBody(t *testing.T) {
	full := replyBytes(t, hostserver.ServerIDRemoteCommand, 1, 0x8003)
	for name, partial := range map[string][]byte{"header": full[:10], "body": full[:21]} {
		t.Run(name, func(t *testing.T) {
			c, server := newPipeConn(t, hostserver.Options{})
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				_, _ = server.Write(partial)
				time.Sleep(20 * time.Millisecond)
				cancel()
			}()
			_, err := c.Receive(ctx)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("Receive error = %v", err)
			}
			wantPoisoned(t, c, err)
		})
	}
}

func TestFullSendWithoutReplyTimesOut(t *testing.T) {
	c, server := newPipeConn(t, hostserver.Options{OperationTimeout: 100 * time.Millisecond})
	serveOne(t, server, nil) // reads the request, never replies
	_, receipt, err := c.Exchange(context.Background(), request(1), expect1)
	want, _ := request(1).Marshal(0)
	if receipt.BytesAccepted != len(want) {
		t.Errorf("receipt = %d; the full request was accepted", receipt.BytesAccepted)
	}
	var ne net.Error
	if !errors.As(err, &ne) || !ne.Timeout() {
		t.Fatalf("error = %v; want a timeout", err)
	}
	wantPoisoned(t, c, err)
}

func TestReceiveRejectsOversizedLengthBeforeAllocating(t *testing.T) {
	c, server := newPipeConn(t, hostserver.Options{MaxFrame: 1024})
	hdr := replyBytes(t, hostserver.ServerIDRemoteCommand, 1, 0x8003)[:hostserver.HeaderLength]
	binary.BigEndian.PutUint32(hdr, 1<<30)
	go server.Write(hdr)
	_, err := c.Receive(context.Background())
	if !errors.Is(err, hostserver.ErrFrameTooLarge) {
		t.Fatalf("error = %v", err)
	}
	wantPoisoned(t, c, err)
}

func TestReceiveRejectsMalformedHeader(t *testing.T) {
	c, server := newPipeConn(t, hostserver.Options{})
	hdr := replyBytes(t, hostserver.ServerIDRemoteCommand, 1, 0x8003)
	binary.BigEndian.PutUint32(hdr, 12) // length below 20
	go server.Write(hdr[:hostserver.HeaderLength])
	_, err := c.Receive(context.Background())
	if !errors.Is(err, hostserver.ErrMalformed) {
		t.Fatalf("error = %v", err)
	}
	wantPoisoned(t, c, err)
}

// trickleConn returns at most one byte per Read and accepts at most one
// byte per Write (with a nil error) to exercise exact I/O loops.
type trickleConn struct{ net.Conn }

func (c trickleConn) Read(b []byte) (int, error) {
	if len(b) > 1 {
		b = b[:1]
	}
	return c.Conn.Read(b)
}

func (c trickleConn) Write(b []byte) (int, error) {
	if len(b) > 1 {
		b = b[:1]
	}
	return c.Conn.Write(b)
}

func TestShortReadsAndWrites(t *testing.T) {
	client, server := net.Pipe()
	c, err := hostserver.NewConn(trickleConn{client}, hostserver.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	defer server.Close()
	sent := serveOne(t, server, replyBytes(t, hostserver.ServerIDRemoteCommand, 1, 0x9003))
	reply, _, err := c.Exchange(context.Background(), request(1), expect1)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := request(1).Marshal(0)
	if got := <-sent; !bytes.Equal(got, want) || reply.Header.RequestID != 0x9003 {
		t.Fatalf("short I/O exchange: sent % X, reply %+v", got, reply)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	c, _ := newPipeConn(t, hostserver.Options{})
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Send(context.Background(), request(1)); !errors.Is(err, hostserver.ErrClosed) {
		t.Errorf("Send after Close = %v", err)
	}
	if _, err := c.Receive(context.Background()); !errors.Is(err, hostserver.ErrClosed) {
		t.Errorf("Receive after Close = %v", err)
	}
}

func TestNewConnRejectsBadOptions(t *testing.T) {
	client, _ := net.Pipe()
	if _, err := hostserver.NewConn(client, hostserver.Options{MaxFrame: 10}); err == nil {
		t.Error("MaxFrame 10 accepted")
	}
	if _, err := hostserver.NewConn(client, hostserver.Options{OperationTimeout: -1}); err == nil {
		t.Error("negative timeout accepted")
	}
}

func TestSendRejectsOversizedFrameWithoutPoisoning(t *testing.T) {
	c, _ := newPipeConn(t, hostserver.Options{MaxFrame: 64})
	_, err := c.Send(context.Background(), hostserver.Frame{Variable: make([]byte, 100)})
	if !errors.Is(err, hostserver.ErrFrameTooLarge) || c.Err() != nil {
		t.Fatalf("Send = %v, Err %v; a local size error must not poison", err, c.Err())
	}
}

type recordingTracer struct {
	mu     sync.Mutex
	events []hostserver.Event
}

func (r *recordingTracer) Trace(e hostserver.Event) {
	r.mu.Lock()
	r.events = append(r.events, e)
	r.mu.Unlock()
}

func TestTracerReportsMetadataOnly(t *testing.T) {
	tr := &recordingTracer{}
	c, server := newPipeConn(t, hostserver.Options{Tracer: tr})
	serveOne(t, server, replyBytes(t, hostserver.ServerIDRemoteCommand, 1, 0x8003))
	if _, _, err := c.Exchange(context.Background(), request(1), expect1); err != nil {
		t.Fatal(err)
	}
	c.Close()
	kinds := []hostserver.EventKind{}
	for _, e := range tr.events {
		kinds = append(kinds, e.Kind)
	}
	want := []hostserver.EventKind{hostserver.EventFrameSent, hostserver.EventFrameReceived, hostserver.EventClosed}
	if len(kinds) != len(want) {
		t.Fatalf("events = %v", kinds)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("events = %v; want %v", kinds, want)
		}
	}
	sent := tr.events[0]
	if sent.RequestID != 0x1003 || sent.CorrelationID != 1 || sent.Bytes != 29 || sent.ServerID != hostserver.ServerIDRemoteCommand {
		t.Errorf("sent event = %+v", sent)
	}
}

func TestConcurrentExchangesSerialize(t *testing.T) {
	client, server := net.Pipe()
	c, err := hostserver.NewConn(client, hostserver.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	defer server.Close()
	// Echo server: answers each request with its own correlation ID.
	go func() {
		for {
			var hb [hostserver.HeaderLength]byte
			if _, err := io.ReadFull(server, hb[:]); err != nil {
				return
			}
			body := make([]byte, binary.BigEndian.Uint32(hb[:])-hostserver.HeaderLength)
			if _, err := io.ReadFull(server, body); err != nil {
				return
			}
			corr := binary.BigEndian.Uint32(hb[12:])
			if _, err := server.Write(replyBytes(t, hostserver.ServerIDRemoteCommand, corr, 0x8003)); err != nil {
				return
			}
		}
	}()
	var corr hostserver.Correlator
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, _ := corr.Next()
			e := expect1
			e.CorrelationID = id
			if _, _, err := c.Exchange(context.Background(), request(id), e); err != nil {
				t.Errorf("exchange %d: %v", id, err)
			}
		}()
	}
	wg.Wait()
	if c.Err() != nil {
		t.Fatalf("Conn poisoned by concurrent use: %v", c.Err())
	}
}

func TestContextDeadlineIsReported(t *testing.T) {
	c, server := newPipeConn(t, hostserver.Options{OperationTimeout: time.Minute})
	serveOne(t, server, nil) // reads the request, never replies
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, _, err := c.Exchange(ctx, request(1), expect1)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v; want context.DeadlineExceeded", err)
	}
	wantPoisoned(t, c, err)
}

func TestPoisonFromAbove(t *testing.T) {
	c, _ := newPipeConn(t, hostserver.Options{})
	cause := errors.New("malformed operation body")
	err := c.Poison(cause)
	if !errors.Is(err, hostserver.ErrPoisoned) || !errors.Is(err, cause) {
		t.Fatalf("Poison = %v", err)
	}
	wantPoisoned(t, c, err)
}

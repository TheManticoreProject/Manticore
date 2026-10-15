package dataqueue_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/dataqueue"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/hosttest"
	"github.com/TheManticoreProject/Manticore/network/ibmi/signon"
)

const (
	testProfile  = "CANARYUSR"
	testPassword = "CANARYPW7"
)

var endpoint = ibminet.Endpoint{Service: ibminet.ServiceDataQueue, Name: "as-dtaq", Host: "ibmi.test", Port: hosttest.DataQueuePort, Source: ibminet.EndpointExplicit}

func newServer() *hosttest.Server {
	return &hosttest.Server{Profile: testProfile, Password: testPassword}
}

func start(t *testing.T, srv *hosttest.Server, cred ibminet.Credential, conn hostserver.Options) (*dataqueue.Session, error) {
	t.Helper()
	c, err := (&hostserver.Dialer{Dial: srv.Dial, Conn: conn}).DialEndpoint(context.Background(), endpoint)
	if err != nil {
		t.Fatal(err)
	}
	return dataqueue.Start(context.Background(), c, cred, ibminet.PasswordFamilyDES, dataqueue.StartOptions{SystemCCSID: 37})
}

func password(t *testing.T, pw string) *ibminet.PasswordCredential {
	t.Helper()
	cred, err := ibminet.NewPasswordCredential(testProfile, []byte(pw))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cred.Destroy)
	return cred
}

func session(t *testing.T, srv *hosttest.Server) *dataqueue.Session {
	t.Helper()
	s, err := start(t, srv, password(t, testPassword), hostserver.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

var (
	fifo  = dataqueue.Queue{Library: "qgpl", Name: "fifo"}
	lifo  = dataqueue.Queue{Library: "QGPL", Name: "LIFO"}
	keyed = dataqueue.Queue{Library: "QGPL", Name: "KEYED"}
)

func TestStartAndClose(t *testing.T) {
	srv := newServer()
	s := session(t, srv)
	if info := s.Info(); info.SystemCCSID != 37 || info.ServerJob != "QZRCSRVS" {
		t.Errorf("info = %+v", info)
	}
	var ids []string
	for _, r := range srv.Requests() {
		if r.ServerID != hostserver.ServerIDDataQueue {
			t.Errorf("request 0x%04X carries server ID 0x%04X", r.RequestID, r.ServerID)
		}
		ids = append(ids, fmt.Sprintf("%04X/%d", r.RequestID, r.CorrelationID))
	}
	if got := strings.Join(ids, " "); got != "7001/0 7002/0 0000/1" {
		t.Errorf("startup = %s", got)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Errorf("second Close = %v", err)
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestStartRejectedCredentialIsNotRetried(t *testing.T) {
	srv := newServer()
	if _, err := start(t, srv, password(t, "WRONGPW1"), hostserver.Options{}); !errors.Is(err, signon.ErrAuthenticationRejected) {
		t.Fatalf("wrong password = %v", err)
	}
	if srv.Count(0x7002) != 1 || srv.Count(0x0000) != 0 {
		t.Errorf("requests = %+v", srv.Requests())
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err) // Start closes the connection on failure
	}
}

func TestStartRejectsAttributesFailure(t *testing.T) {
	srv := newServer()
	srv.DataQueueReply = func(f hostserver.Frame) (hostserver.Frame, bool) {
		return hostserver.Frame{Header: hostserver.Header{ServerID: hostserver.ServerIDDataQueue, RequestID: 0x8002, CorrelationID: f.Header.CorrelationID}, Variable: []byte{0xF0, 0x08}}, true
	}
	var e *dataqueue.Error
	if _, err := start(t, srv, password(t, testPassword), hostserver.Options{}); !errors.As(err, &e) || e.ReturnCode != 0xF008 {
		t.Fatalf("Start = %v", err)
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestStartWithProfileToken(t *testing.T) {
	srv := newServer()
	token, err := signon.GenerateProfileToken(context.Background(),
		ibminet.Endpoint{Service: ibminet.ServiceSignon, Name: "as-signon", Host: "ibmi.test", Port: hosttest.SignonPort, Source: ibminet.EndpointExplicit},
		password(t, testPassword), signon.TokenRequest{Type: ibminet.TokenMultipleUse}, signon.Options{Dialer: &hostserver.Dialer{Dial: srv.Dial}})
	if err != nil {
		t.Fatal(err)
	}
	defer token.Destroy()
	s, err := start(t, srv, token, hostserver.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, r := range srv.Requests() {
		if r.ServerID == hostserver.ServerIDDataQueue && r.RequestID == 0x7002 && (r.CredentialType != 0x02 || len(r.Substitute) != 32) {
			t.Errorf("0x7002 credential type 0x%02X, %d bytes", r.CredentialType, len(r.Substitute))
		}
	}
}

func TestQueueLifecycle(t *testing.T) {
	ctx := context.Background()
	srv := newServer()
	s := session(t, srv)

	if _, err := s.Attributes(ctx, fifo); !errors.Is(err, dataqueue.ErrQueueNotFound) || !errors.Is(err, dataqueue.ErrServerRejected) {
		t.Fatalf("missing queue = %v", err)
	} else if strings.Contains(err.Error(), "not found.") {
		t.Errorf("error %q contains the server's message text", err)
	}
	if err := s.Create(ctx, fifo, dataqueue.CreateOptions{MaxEntryLength: 16, Kind: dataqueue.FIFO, Description: "first"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(ctx, fifo, dataqueue.CreateOptions{MaxEntryLength: 16, Kind: dataqueue.FIFO}); !errors.Is(err, dataqueue.ErrAlreadyExists) {
		t.Errorf("second create = %v", err)
	}
	a, err := s.Attributes(ctx, fifo)
	if err != nil {
		t.Fatal(err)
	}
	if *a != (dataqueue.Attributes{MaxEntryLength: 16, Kind: dataqueue.FIFO, Description: "first"}) {
		t.Errorf("attributes = %+v", a)
	}

	if e, err := s.Read(ctx, fifo, 0); e != nil || err != nil {
		t.Fatalf("empty read = %+v, %v", e, err)
	}
	for _, d := range []string{"one", "two", "three"} {
		if err := s.Write(ctx, fifo, []byte(d)); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Write(ctx, fifo, make([]byte, 17)); !errors.Is(err, dataqueue.ErrServerRejected) {
		t.Errorf("oversized write = %v", err)
	}
	if e, err := s.Peek(ctx, fifo, 0); err != nil || string(e.Data) != "one" || e.Sender != nil || e.Key != nil {
		t.Fatalf("peek = %+v, %v", e, err)
	}
	for _, want := range []string{"one", "two"} {
		if e, err := s.Read(ctx, fifo, time.Second); err != nil || string(e.Data) != want {
			t.Fatalf("read = %+v, %v; want %s", e, err, want)
		}
	}
	if err := s.Clear(ctx, fifo); err != nil {
		t.Fatal(err)
	}
	if e, err := s.Read(ctx, fifo, 0); e != nil || err != nil {
		t.Errorf("read after clear = %+v, %v", e, err)
	}
	if _, err := s.ReadKey(ctx, fifo, dataqueue.SearchEqual, []byte("K"), 0); !errors.Is(err, dataqueue.ErrWrongQueueKind) {
		t.Errorf("keyed read of a FIFO queue = %v", err)
	}
	if err := s.Delete(ctx, fifo); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, fifo); !errors.Is(err, dataqueue.ErrQueueNotFound) {
		t.Errorf("second delete = %v", err)
	}

	if err := s.Create(ctx, lifo, dataqueue.CreateOptions{MaxEntryLength: 8, Kind: dataqueue.LIFO, Authority: dataqueue.AuthorityUse}); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{"a", "b"} {
		if err := s.Write(ctx, lifo, []byte(d)); err != nil {
			t.Fatal(err)
		}
	}
	if e, err := s.Read(ctx, lifo, 0); err != nil || string(e.Data) != "b" {
		t.Errorf("LIFO read = %+v, %v", e, err)
	}
	if err := s.Err(); err != nil {
		t.Fatalf("session unusable after server errors: %v", err)
	}
}

func TestKeyedSearches(t *testing.T) {
	ctx := context.Background()
	srv := newServer()
	s := session(t, srv)
	if err := s.Create(ctx, keyed, dataqueue.CreateOptions{MaxEntryLength: 8, Kind: dataqueue.Keyed, KeyLength: 2, SaveSenderInfo: true}); err != nil {
		t.Fatal(err)
	}
	if a, err := s.Attributes(ctx, keyed); err != nil || a.Kind != dataqueue.Keyed || a.KeyLength != 2 || !a.SaveSenderInfo {
		t.Fatalf("attributes = %+v, %v", a, err)
	}
	for _, k := range []string{"20", "10", "30"} {
		if err := s.WriteKey(ctx, keyed, []byte(k), []byte("e"+k)); err != nil {
			t.Fatal(err)
		}
	}
	for _, tt := range []struct {
		op   dataqueue.Search
		want string
	}{
		{dataqueue.SearchEqual, "10"},
		{dataqueue.SearchNotEqual, "20"},
		{dataqueue.SearchLess, ""},
		{dataqueue.SearchLessOrEqual, "10"},
		{dataqueue.SearchGreater, "20"},
		{dataqueue.SearchGreaterOrEqual, "20"},
	} {
		e, err := s.PeekKey(ctx, keyed, tt.op, []byte("10"), 0)
		if err != nil {
			t.Fatalf("%s: %v", tt.op, err)
		}
		if got := ""; e != nil {
			got = string(e.Key)
			if got != tt.want || string(e.Data) != "e"+got {
				t.Errorf("%s 10 = %+v; want key %q", tt.op, e, tt.want)
			}
		} else if tt.want != "" {
			t.Errorf("%s 10 found nothing; want %q", tt.op, tt.want)
		}
	}
	e, err := s.ReadKey(ctx, keyed, dataqueue.SearchEqual, []byte("30"), 0)
	if err != nil || e == nil || e.Sender == nil {
		t.Fatalf("read EQ 30 = %+v, %v", e, err)
	}
	if want := (dataqueue.Sender{JobName: "QZRCSRVS", JobUser: "QUSER", JobNumber: "000042", CurrentUser: testProfile}); *e.Sender != want || e.Sender.Job() != "000042/QUSER/QZRCSRVS" {
		t.Errorf("sender = %+v", *e.Sender)
	}
	if err := s.ClearKey(ctx, keyed, []byte("10")); err != nil {
		t.Fatal(err)
	}
	if e, err := s.ReadKey(ctx, keyed, dataqueue.SearchLessOrEqual, []byte("99"), 0); err != nil || string(e.Key) != "20" {
		t.Errorf("after clear key = %+v, %v", e, err)
	}
	if _, err := s.Read(ctx, keyed, 0); !errors.Is(err, dataqueue.ErrWrongQueueKind) {
		t.Errorf("sequential read of a keyed queue = %v", err)
	}
}

func TestValidationSendsNothing(t *testing.T) {
	ctx := context.Background()
	srv := newServer()
	s := session(t, srv)
	before := len(srv.Requests())
	for name, call := range map[string]func() error{
		"empty name":       func() error { _, err := s.Attributes(ctx, dataqueue.Queue{Library: "QGPL"}); return err },
		"long name":        func() error { return s.Delete(ctx, dataqueue.Queue{Library: "QGPL", Name: "ABCDEFGHIJK"}) },
		"digit first":      func() error { return s.Delete(ctx, dataqueue.Queue{Library: "1LIB", Name: "Q"}) },
		"special value":    func() error { return s.Delete(ctx, dataqueue.Queue{Library: "*LIBL", Name: "Q"}) },
		"path separator":   func() error { return s.Delete(ctx, dataqueue.Queue{Library: "QGPL", Name: "A/B"}) },
		"negative wait":    func() error { _, err := s.Read(ctx, fifo, -2*time.Second); return err },
		"fractional wait":  func() error { _, err := s.Read(ctx, fifo, 1500*time.Millisecond); return err },
		"undefined search": func() error { _, err := s.ReadKey(ctx, keyed, 9, []byte("K"), 0); return err },
		"empty key":        func() error { _, err := s.ReadKey(ctx, keyed, dataqueue.SearchEqual, nil, 0); return err },
		"empty write key":  func() error { return s.WriteKey(ctx, keyed, nil, []byte("x")) },
		"empty clear key":  func() error { return s.ClearKey(ctx, keyed, nil) },
		"long key":         func() error { return s.WriteKey(ctx, keyed, make([]byte, 257), []byte("x")) },
		"long entry":       func() error { return s.Write(ctx, fifo, make([]byte, 64513)) },
		"no kind":          func() error { return s.Create(ctx, fifo, dataqueue.CreateOptions{MaxEntryLength: 1}) },
		"zero max entry":   func() error { return s.Create(ctx, fifo, dataqueue.CreateOptions{Kind: dataqueue.FIFO}) },
		"keyed without key length": func() error {
			return s.Create(ctx, keyed, dataqueue.CreateOptions{MaxEntryLength: 1, Kind: dataqueue.Keyed})
		},
		"authority": func() error {
			return s.Create(ctx, fifo, dataqueue.CreateOptions{MaxEntryLength: 1, Kind: dataqueue.FIFO, Authority: 9})
		},
		"long description": func() error {
			return s.Create(ctx, fifo, dataqueue.CreateOptions{MaxEntryLength: 1, Kind: dataqueue.FIFO, Description: strings.Repeat("d", 51)})
		},
	} {
		if err := call(); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if n := len(srv.Requests()); n != before {
		t.Errorf("validation failures sent %d requests", n-before)
	}
	if !errors.Is(s.Delete(ctx, dataqueue.Queue{Library: "QGPL", Name: "A.B"}), dataqueue.ErrQueueNotFound) {
		t.Error("a valid name with a period was rejected locally")
	}
}

func TestLostReplyIsUncertainAndNeverRetried(t *testing.T) {
	srv := newServer()
	srv.SilentOn = 0x0005
	s := session(t, srv)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err := s.Write(ctx, fifo, []byte("x"))
	var oe *dataqueue.OutcomeError
	if !errors.Is(err, dataqueue.ErrUncertainOutcome) || !errors.As(err, &oe) || !oe.MayHaveExecuted || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("lost write reply = %v", err)
	}
	if s.Err() == nil {
		t.Fatal("the session stays usable after a lost reply")
	}
	if err := s.Delete(context.Background(), fifo); err == nil || errors.Is(err, dataqueue.ErrUncertainOutcome) {
		t.Errorf("operation on a poisoned session = %v; want a not-started failure", err)
	}
	if n := srv.Count(0x0005); n != 1 {
		t.Errorf("write sent %d times", n)
	}
}

func TestMalformedReplyPoisons(t *testing.T) {
	srv := newServer()
	srv.DataQueueReply = func(f hostserver.Frame) (hostserver.Frame, bool) {
		if f.Header.RequestID != 0x0002 {
			return hostserver.Frame{}, false
		}
		body := append(append([]byte{0, 0}, bytes.Repeat([]byte{0x40}, 36)...), 0, 0, 0, 0x40, 0x50, 0x01)
		return hostserver.Frame{Header: hostserver.Header{ServerID: hostserver.ServerIDDataQueue, RequestID: 0x8003, CorrelationID: f.Header.CorrelationID}, Variable: body}, true
	}
	s := session(t, srv)
	_, err := s.Read(context.Background(), fifo, 0)
	if !errors.Is(err, hostserver.ErrMalformed) || !errors.Is(err, dataqueue.ErrUncertainOutcome) {
		t.Fatalf("malformed read reply = %v", err)
	}
	if !errors.Is(s.Err(), hostserver.ErrPoisoned) {
		t.Errorf("Err() = %v", s.Err())
	}
}

func TestReadWaitOutlivesOperationTimeout(t *testing.T) {
	srv := newServer()
	srv.DataQueueReply = func(f hostserver.Frame) (hostserver.Frame, bool) {
		if f.Header.RequestID == 0x0002 {
			time.Sleep(300 * time.Millisecond) // the server holds the read
		}
		return hostserver.Frame{}, false
	}
	s, err := start(t, srv, password(t, testPassword), hostserver.Options{OperationTimeout: 100 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.Create(context.Background(), fifo, dataqueue.CreateOptions{MaxEntryLength: 1, Kind: dataqueue.FIFO}); err != nil {
		t.Fatal(err)
	}
	if e, err := s.Read(context.Background(), fifo, time.Second); e != nil || err != nil {
		t.Fatalf("1s read = %+v, %v", e, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	if _, err := s.Read(ctx, fifo, dataqueue.WaitForever); !errors.Is(err, context.DeadlineExceeded) || !errors.Is(err, dataqueue.ErrUncertainOutcome) {
		t.Errorf("cancelled indefinite read = %v", err)
	}
}

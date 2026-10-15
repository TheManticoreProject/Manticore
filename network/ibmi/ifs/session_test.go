package ifs_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/ifs"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/hosttest"
	"github.com/TheManticoreProject/Manticore/network/ibmi/signon"
)

const (
	testProfile  = "CANARYUSR"
	testPassword = "CANARYPW7"
)

var endpoint = ibminet.Endpoint{Service: ibminet.ServiceFile, Name: "as-file", Host: "ibmi.test", Port: hosttest.FilePort, Source: ibminet.EndpointExplicit}

// content is deterministic, non-repeating file data.
func content(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i*7 + i/251)
	}
	return b
}

func newServer() *hosttest.Server {
	return &hosttest.Server{Profile: testProfile, Password: testPassword, Files: map[string][]byte{
		"/home/canary/small.txt": []byte("hello, IFS\n"),
		"/big.bin":               content(300_000),
		"/empty":                 {},
		"/QIBM/ProdData/é.txt":   []byte("non-ASCII path"),
	}}
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

func start(t *testing.T, srv *hosttest.Server, cred ibminet.Credential, opts ifs.StartOptions) (*ifs.Session, error) {
	t.Helper()
	c, err := (&hostserver.Dialer{Dial: srv.Dial}).DialEndpoint(context.Background(), endpoint)
	if err != nil {
		t.Fatal(err)
	}
	return ifs.Start(context.Background(), c, cred, ibminet.PasswordFamilyDES, opts)
}

func session(t *testing.T, srv *hosttest.Server, opts ifs.StartOptions) *ifs.Session {
	t.Helper()
	s, err := start(t, srv, password(t, testPassword), opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestStartAndClose(t *testing.T) {
	srv := newServer()
	s := session(t, srv, ifs.StartOptions{SystemCCSID: 37})
	if info := s.Info(); info != (ifs.Info{DataStreamLevel: 12, MaxBlock: 0x40000000, PathCCSID: 1200, ServerJob: "QZRCSRVS"}) {
		t.Errorf("info = %+v", info)
	}
	var ids []string
	for _, r := range srv.Requests() {
		if r.ServerID != hostserver.ServerIDFile {
			t.Errorf("request 0x%04X carries server ID 0x%04X", r.RequestID, r.ServerID)
		}
		ids = append(ids, fmt.Sprintf("%04X/%d", r.RequestID, r.CorrelationID))
	}
	if got := strings.Join(ids, " "); got != "7001/0 7002/0 0016/1" {
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

func TestStartFailures(t *testing.T) {
	srv := newServer()
	if _, err := start(t, srv, password(t, "WRONGPW1"), ifs.StartOptions{}); !errors.Is(err, signon.ErrAuthenticationRejected) {
		t.Errorf("wrong password = %v", err)
	}
	if srv.Count(0x7002) != 1 || srv.Count(0x0016) != 0 {
		t.Errorf("requests = %+v", srv.Requests())
	}
	for name, mutate := range map[string]func(*hosttest.Server){
		"level 16":    func(s *hosttest.Server) { s.FileLevel = 16 },
		"CCSID 61952": func(s *hosttest.Server) { s.FilePathCCSID = 61952 },
	} {
		srv := newServer()
		mutate(srv)
		if _, err := start(t, srv, password(t, testPassword), ifs.StartOptions{}); !errors.Is(err, ifs.ErrUnsupported) {
			t.Errorf("%s = %v", name, err)
		}
		if err := srv.WaitClosed(2 * time.Second); err != nil {
			t.Errorf("%s: %v", name, err) // Start closes the connection
		}
	}
	srv = newServer()
	srv.FilePathCCSID = 13488
	if _, err := start(t, srv, password(t, testPassword), ifs.StartOptions{}); err != nil {
		t.Errorf("CCSID 13488 = %v", err)
	}
	for _, opts := range []ifs.StartOptions{{MaxChunk: -1}, {MaxChunk: ifs.MaxMaxChunk + 1}, {MaxChain: -1}, {SystemCCSID: 1}} {
		if _, err := start(t, newServer(), password(t, testPassword), opts); err == nil {
			t.Errorf("options %+v accepted", opts)
		}
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
	s, err := start(t, srv, token, ifs.StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if data, err := s.ReadFile(context.Background(), "/home/canary/small.txt", 100); err != nil || string(data) != "hello, IFS\n" {
		t.Errorf("read with token = %q, %v", data, err)
	}
}

func TestOpenReadClose(t *testing.T) {
	ctx := context.Background()
	srv := newServer()
	s := session(t, srv, ifs.StartOptions{})
	f, err := s.Open(ctx, "/home/canary/small.txt")
	if err != nil {
		t.Fatal(err)
	}
	if info := f.Info(); info.Size != 11 || info.DataCCSID != 819 || !info.Modified.Equal(time.Unix(1700000000, 0)) {
		t.Errorf("info = %+v", info)
	}
	if data, err := f.ReadAt(ctx, 0, 5); err != nil || string(data) != "hello" {
		t.Errorf("read 0:5 = %q, %v", data, err)
	}
	// A short read at the tail is not EOF; a read at the end is.
	if data, err := f.ReadAt(ctx, 7, 100); err != nil || string(data) != "IFS\n" {
		t.Errorf("tail = %q, %v", data, err)
	}
	if data, err := f.ReadAt(ctx, 11, 100); !errors.Is(err, io.EOF) || len(data) != 0 {
		t.Errorf("at end = %q, %v", data, err)
	}
	if err := f.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(ctx); err != nil {
		t.Errorf("second Close = %v", err)
	}
	if srv.Count(0x0009) != 1 {
		t.Errorf("close sent %d times", srv.Count(0x0009))
	}
	if _, err := f.ReadAt(ctx, 0, 1); !errors.Is(err, ifs.ErrFileClosed) {
		t.Errorf("read after close = %v", err)
	}

	if f, err = s.Open(ctx, "/QIBM/ProdData/é.txt"); err != nil {
		t.Errorf("non-ASCII path: %v", err)
	} else {
		f.Close(ctx)
	}
	_, err = s.Open(ctx, "/missing")
	var e *ifs.Error
	if !errors.Is(err, ifs.ErrNotFound) || !errors.As(err, &e) || e.ReturnCode != 2 || strings.Contains(err.Error(), "missing") {
		t.Errorf("missing file = %v", err)
	}
	if _, err := s.Open(ctx, "/Home/Canary/Small.txt"); !errors.Is(err, ifs.ErrNotFound) {
		t.Errorf("path case was changed: %v", err)
	}
	if err := s.Err(); err != nil {
		t.Fatalf("session unusable after a server error: %v", err)
	}
}

func TestReadFile(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name  string
		chunk int
		eofRC bool
	}{{"single packets", 0, false}, {"chained packets", 7000, false}, {"return-code EOF", 0, true}, {"chained, return-code EOF", 7000, true}} {
		t.Run(tc.name, func(t *testing.T) {
			srv := newServer()
			srv.FileChunk, srv.FileEOFReturnCode = tc.chunk, tc.eofRC
			s := session(t, srv, ifs.StartOptions{MaxChunk: 64 << 10})
			data, err := s.ReadFile(ctx, "/big.bin", 300_000)
			if err != nil || !bytes.Equal(data, content(300_000)) {
				t.Fatalf("ReadFile = %d bytes, %v", len(data), err)
			}
			if data, err := s.ReadFile(ctx, "/empty", 1); err != nil || len(data) != 0 {
				t.Errorf("empty file = %q, %v", data, err)
			}
			if _, err := s.ReadFile(ctx, "/big.bin", 299_999); !errors.Is(err, ifs.ErrTooLarge) {
				t.Errorf("over the limit = %v", err)
			}
			if srv.Count(0x0002) != srv.Count(0x0009) {
				t.Errorf("%d opens, %d closes", srv.Count(0x0002), srv.Count(0x0009))
			}
			if err := s.Err(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// The size reported at open is not trusted: a file that grew is caught by
// reading at most limit+1 bytes.
func TestReadFileStopsAtLimitPlusOne(t *testing.T) {
	srv := newServer()
	var requested []int
	srv.FileReply = func(f hostserver.Frame) ([]hostserver.Frame, bool) {
		switch f.Header.RequestID {
		case 0x0002:
			body := make([]byte, 69)
			body[5] = 0x01 // handle 1; reported size 0
			return []hostserver.Frame{{Header: hostserver.Header{ServerID: hostserver.ServerIDFile, RequestID: 0x8002, CorrelationID: f.Header.CorrelationID}, Variable: body}}, true
		case 0x0003:
			n := int(f.Template[14])<<24 | int(f.Template[15])<<16 | int(f.Template[16])<<8 | int(f.Template[17])
			requested = append(requested, n)
			body := []byte{0, 0, 0xFF, 0xFF}
			body, _ = hostserver.AppendField(body, 5, make([]byte, n))
			return []hostserver.Frame{{Header: hostserver.Header{ServerID: hostserver.ServerIDFile, RequestID: 0x8003, CorrelationID: f.Header.CorrelationID}, Variable: body}}, true
		}
		return nil, false
	}
	s := session(t, srv, ifs.StartOptions{})
	if _, err := s.ReadFile(context.Background(), "/growing", 10); !errors.Is(err, ifs.ErrTooLarge) {
		t.Fatalf("growing file = %v", err)
	}
	if len(requested) != 1 || requested[0] != 11 {
		t.Errorf("read requests = %v; want one of 11 bytes", requested)
	}
}

func TestChainKeepsBytesBeforeEOF(t *testing.T) {
	srv := newServer()
	srv.FileReply = func(f hostserver.Frame) ([]hostserver.Frame, bool) {
		if f.Header.RequestID != 0x0003 {
			return nil, false
		}
		mk := func(chain byte, data []byte) hostserver.Frame {
			body := []byte{0, chain, 0xFF, 0xFF}
			body, _ = hostserver.AppendField(body, 5, data)
			return hostserver.Frame{Header: hostserver.Header{ServerID: hostserver.ServerIDFile, RequestID: 0x8003, CorrelationID: f.Header.CorrelationID}, Variable: body}
		}
		eof := hostserver.Frame{Header: hostserver.Header{ServerID: hostserver.ServerIDFile, RequestID: 0x8001, CorrelationID: f.Header.CorrelationID}, Variable: []byte{0, 0, 0, 22}}
		return []hostserver.Frame{mk(1, []byte("ab")), mk(1, []byte("cd")), eof}, true
	}
	s := session(t, srv, ifs.StartOptions{})
	f, err := s.Open(context.Background(), "/home/canary/small.txt")
	if err != nil {
		t.Fatal(err)
	}
	data, err := f.ReadAt(context.Background(), 0, 100)
	if !errors.Is(err, io.EOF) || string(data) != "abcd" {
		t.Fatalf("chain ending in EOF = %q, %v", data, err)
	}
	if err := s.Err(); err != nil {
		t.Fatalf("the chain was not fully consumed: %v", err)
	}
}

func TestMalformedChainsPoison(t *testing.T) {
	mk := func(corr uint32, id uint16, chain byte, data []byte) hostserver.Frame {
		body := []byte{0, chain, 0xFF, 0xFF}
		body, _ = hostserver.AppendField(body, 5, data)
		return hostserver.Frame{Header: hostserver.Header{ServerID: hostserver.ServerIDFile, RequestID: id, CorrelationID: corr}, Variable: body}
	}
	for name, replies := range map[string]func(corr uint32) []hostserver.Frame{
		"data after EOF": func(c uint32) []hostserver.Frame {
			return []hostserver.Frame{mk(c, 0x8003, 1, nil), mk(c, 0x8003, 0, []byte("x"))}
		},
		"more than requested": func(c uint32) []hostserver.Frame { return []hostserver.Frame{mk(c, 0x8003, 0, make([]byte, 102))} },
		"wrong chain correlation": func(c uint32) []hostserver.Frame {
			return []hostserver.Frame{mk(c, 0x8003, 1, []byte("a")), mk(c+1, 0x8003, 0, []byte("b"))}
		},
		"wrong chain reply ID": func(c uint32) []hostserver.Frame {
			return []hostserver.Frame{mk(c, 0x8003, 1, []byte("a")), mk(c, 0x8002, 0, []byte("b"))}
		},
		"chain too long": func(c uint32) []hostserver.Frame {
			return []hostserver.Frame{mk(c, 0x8003, 1, []byte("a")), mk(c, 0x8003, 1, []byte("b")), mk(c, 0x8003, 0, []byte("c"))}
		},
		"no data, no EOF": func(c uint32) []hostserver.Frame {
			return []hostserver.Frame{{Header: hostserver.Header{ServerID: hostserver.ServerIDFile, RequestID: 0x8001, CorrelationID: c}, Variable: []byte{0, 0, 0, 0}}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			srv := newServer()
			srv.FileReply = func(f hostserver.Frame) ([]hostserver.Frame, bool) {
				if f.Header.RequestID != 0x0003 {
					return nil, false
				}
				return replies(f.Header.CorrelationID), true
			}
			s := session(t, srv, ifs.StartOptions{MaxChain: 2})
			_, err := s.ReadFile(context.Background(), "/home/canary/small.txt", 100) // asks for 101 bytes
			if !errors.Is(err, ifs.ErrUncertainOutcome) || !errors.Is(s.Err(), hostserver.ErrPoisoned) {
				t.Errorf("ReadFile = %v, Err() = %v", err, s.Err())
			}
		})
	}
}

func TestReadErrorsAndValidation(t *testing.T) {
	ctx := context.Background()
	srv := newServer()
	s := session(t, srv, ifs.StartOptions{MaxChunk: 100})
	f, err := s.Open(ctx, "/home/canary/small.txt")
	if err != nil {
		t.Fatal(err)
	}
	before := len(srv.Requests())
	for _, tt := range []struct {
		off int64
		n   int
	}{{0, 0}, {0, 101}, {-1, 1}, {1 << 31, 1}} {
		if _, err := f.ReadAt(ctx, tt.off, tt.n); err == nil {
			t.Errorf("read %d at %d accepted", tt.n, tt.off)
		}
	}
	for _, p := range []string{"", "/a\x00b", "/bad\xff"} {
		if _, err := s.Open(ctx, p); err == nil {
			t.Errorf("path %q accepted", p)
		}
	}
	for _, limit := range []int{0, -1} {
		if _, err := s.ReadFile(ctx, "/home/canary/small.txt", limit); err == nil {
			t.Errorf("limit %d accepted", limit)
		}
	}
	if n := len(srv.Requests()); n != before {
		t.Errorf("validation failures sent %d requests", n-before)
	}
	// A handle the server no longer knows: return code 6, session intact.
	srv.FileReply = func(fr hostserver.Frame) ([]hostserver.Frame, bool) {
		if fr.Header.RequestID != 0x0003 {
			return nil, false
		}
		return []hostserver.Frame{{Header: hostserver.Header{ServerID: hostserver.ServerIDFile, RequestID: 0x8001, CorrelationID: fr.Header.CorrelationID}, Variable: []byte{0, 0, 0, 6}}}, true
	}
	var e *ifs.Error
	if _, err := f.ReadAt(ctx, 0, 1); !errors.As(err, &e) || e.ReturnCode != 6 || s.Err() != nil {
		t.Errorf("invalid handle = %v, Err() = %v", err, s.Err())
	}
}

func TestLostReplyAndClosedSession(t *testing.T) {
	srv := newServer()
	srv.SilentOn = 0x0003
	s := session(t, srv, ifs.StartOptions{})
	f, err := s.Open(context.Background(), "/home/canary/small.txt")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := f.ReadAt(ctx, 0, 1); !errors.Is(err, ifs.ErrUncertainOutcome) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("lost read reply = %v", err)
	}
	// The handle died with the poisoned connection: Close sends nothing.
	if err := f.Close(context.Background()); err != nil || srv.Count(0x0009) != 0 {
		t.Errorf("Close on a poisoned session = %v, %d close requests", err, srv.Count(0x0009))
	}
	if _, err := s.Open(context.Background(), "/home/canary/small.txt"); err == nil {
		t.Error("Open on a poisoned session succeeded")
	}
}

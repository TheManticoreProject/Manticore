package signon_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/auth"
	"github.com/TheManticoreProject/Manticore/network/ibmi/signon"
	"github.com/TheManticoreProject/Manticore/network/ibmi/signon/message"
)

// Synthetic credentials only; the canary strings let redaction tests detect
// leaks.
const (
	testProfile  = "CANARYUSR"
	testPassword = "CANARYPW7"
)

var serverSeed = [8]byte{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}

// fakeSignon scripts one Sign-on server connection. It verifies the
// substitute it receives against its own copy of the credential.
type fakeSignon struct {
	passwordLevel  *byte
	version        uint32
	attrRC, infoRC uint32
	omitCCSID      bool
	malformedF003  bool
	wrongReplyID   bool
	silent         bool // never answer 0x7003

	mu          sync.Mutex
	requests    []uint16
	clientSeeds [][8]byte
	substitutes [][]byte
	closed      bool
	done        chan struct{}
}

func level(b byte) *byte { return &b }

func (s *fakeSignon) record(id uint16) {
	s.mu.Lock()
	s.requests = append(s.requests, id)
	s.mu.Unlock()
}

func (s *fakeSignon) count(id uint16) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, r := range s.requests {
		if r == id {
			n++
		}
	}
	return n
}

func readFrame(c net.Conn) (hostserver.Frame, error) {
	var hb [hostserver.HeaderLength]byte
	if _, err := io.ReadFull(c, hb[:]); err != nil {
		return hostserver.Frame{}, err
	}
	b := make([]byte, binary.BigEndian.Uint32(hb[:]))
	copy(b, hb[:])
	if _, err := io.ReadFull(c, b[hostserver.HeaderLength:]); err != nil {
		return hostserver.Frame{}, err
	}
	return hostserver.UnmarshalFrame(b, 0)
}

func writeReply(c net.Conn, id uint16, rc uint32, fields ...hostserver.Field) {
	var v []byte
	for _, f := range fields {
		v, _ = hostserver.AppendField(v, f.CodePoint, f.Value)
	}
	b, _ := hostserver.Frame{
		Header:   hostserver.Header{ServerID: hostserver.ServerIDSignon, RequestID: id},
		Template: binary.BigEndian.AppendUint32(nil, rc),
		Variable: v,
	}.Marshal(0)
	_, _ = c.Write(b)
}

func (s *fakeSignon) serve(c net.Conn) {
	defer close(s.done)
	defer c.Close()
	expected, _ := auth.NewPasswordCredential(testProfile, []byte(testPassword))
	var clientSeed [8]byte
	for {
		f, err := readFrame(c)
		if err != nil {
			s.mu.Lock()
			s.closed = true // the client closed the socket
			s.mu.Unlock()
			return
		}
		s.record(f.Header.RequestID)
		switch f.Header.RequestID {
		case message.RequestExchangeAttributes:
			fields, _ := hostserver.ParseFields(f.Variable, hostserver.FieldPolicy{})
			seed, _ := fields.Get(message.CPSeed)
			copy(clientSeed[:], seed)
			s.mu.Lock()
			s.clientSeeds = append(s.clientSeeds, clientSeed)
			s.mu.Unlock()
			if s.silent {
				continue
			}
			id := message.ReplyExchangeAttributes
			if s.wrongReplyID {
				id = message.ReplyRetrieveInfo
			}
			if s.malformedF003 {
				writeReply(c, id, 0, hostserver.Field{CodePoint: message.CPSeed, Value: []byte{1, 2, 3}})
				continue
			}
			fields2 := []hostserver.Field{
				{CodePoint: message.CPVersion, Value: binary.BigEndian.AppendUint32(nil, s.version)},
				{CodePoint: message.CPDataStreamLevel, Value: []byte{0, 2}},
				{CodePoint: message.CPSeed, Value: serverSeed[:]},
			}
			if s.passwordLevel != nil {
				fields2 = append(fields2, hostserver.Field{CodePoint: message.CPPasswordLevel, Value: []byte{*s.passwordLevel}})
			}
			writeReply(c, id, s.attrRC, fields2...)
		case message.RequestRetrieveInfo:
			fields, _ := hostserver.ParseFields(f.Variable, hostserver.FieldPolicy{})
			got, _ := fields.Get(message.CPPasswordSubstitute)
			s.mu.Lock()
			s.substitutes = append(s.substitutes, append([]byte(nil), got...))
			s.mu.Unlock()
			family := auth.FamilyDES
			if f.Template[0] == message.CredentialTypeSHA1 {
				family = auth.FamilySHA1
			}
			want, _ := auth.Derive(expected, family, clientSeed, serverSeed)
			rc := s.infoRC
			if rc == 0 && !bytes.Equal(got, want.Value) {
				rc = 0x0003000B // password incorrect
			}
			var out []hostserver.Field
			if !s.omitCCSID {
				out = append(out, hostserver.Field{CodePoint: message.CPServerCCSID, Value: []byte{0, 0, 0, 37}})
			}
			writeReply(c, message.ReplyRetrieveInfo, rc, out...)
		case message.RequestEndServer:
			// one-way; keep reading until the client closes
		}
	}
}

func run(t *testing.T, s *fakeSignon, password string, ctx context.Context) (*signon.Info, error) {
	t.Helper()
	if s.version == 0 {
		s.version = 0x00050400
	}
	s.done = make(chan struct{})
	client, server := net.Pipe()
	go s.serve(server)
	dialer := &hostserver.Dialer{Dial: func(context.Context, string, string) (net.Conn, error) { return client, nil }}
	cred, err := ibminet.NewPasswordCredential(testProfile, []byte(password))
	if err != nil {
		t.Fatal(err)
	}
	defer cred.Destroy()
	ep := ibminet.Endpoint{Service: ibminet.ServiceSignon, Name: "as-signon", Host: "ibmi.test", Port: 8476, Source: ibminet.EndpointExplicit}
	info, err := signon.Validate(ctx, ep, cred, signon.Options{Dialer: dialer, CleanupTimeout: time.Second})
	select {
	case <-s.done:
	case <-time.After(5 * time.Second):
		t.Fatal("Validate returned without closing the Sign-on socket")
	}
	if !s.closed {
		t.Fatal("Sign-on socket left open")
	}
	return info, err
}

func TestValidateDES(t *testing.T) {
	s := &fakeSignon{passwordLevel: level(0)}
	info, err := run(t, s, testPassword, context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.PasswordFamily != ibminet.PasswordFamilyDES || !info.PasswordLevelReported || info.SystemCCSID != 37 ||
		info.ServerVersion != 0x00050400 || info.VRM.String() != "V5R4M0" || info.ServerLevel != 2 || len(info.Warnings) != 0 {
		t.Errorf("info = %+v", info)
	}
	if len(s.substitutes) != 1 || len(s.substitutes[0]) != 8 {
		t.Errorf("substitutes = %x", s.substitutes)
	}
	if s.count(message.RequestEndServer) != 1 {
		t.Errorf("requests = %04x; want one 0x7006", s.requests)
	}
}

func TestValidateSHA1(t *testing.T) {
	for _, lvl := range []byte{2, 3} {
		s := &fakeSignon{passwordLevel: level(lvl)}
		info, err := run(t, s, testPassword, context.Background())
		if err != nil {
			t.Fatalf("level %d: %v", lvl, err)
		}
		if info.PasswordFamily != ibminet.PasswordFamilySHA1 || info.PasswordLevel != lvl || len(s.substitutes[0]) != 20 {
			t.Errorf("level %d: info %+v, substitute %d bytes", lvl, info, len(s.substitutes[0]))
		}
	}
}

func TestValidateMissingPasswordLevelAssumesDES(t *testing.T) {
	info, err := run(t, &fakeSignon{}, testPassword, context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.PasswordFamily != ibminet.PasswordFamilyDES || info.PasswordLevelReported || len(info.Warnings) != 1 {
		t.Errorf("info = %+v", info)
	}
}

func TestValidateUndecodableVersionWarns(t *testing.T) {
	info, err := run(t, &fakeSignon{passwordLevel: level(0), version: 0x00000A00}, testPassword, context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !info.VRM.IsZero() || len(info.Warnings) != 1 || info.ServerVersion != 0x00000A00 {
		t.Errorf("info = %+v", info)
	}
}

func TestWrongPasswordIsRejectedOnce(t *testing.T) {
	s := &fakeSignon{passwordLevel: level(2)}
	_, err := run(t, s, "WRONGPW1", context.Background())
	var se *signon.Error
	if !errors.Is(err, signon.ErrAuthenticationRejected) || !errors.As(err, &se) || se.Reason() != signon.ReasonPasswordIncorrect {
		t.Fatalf("error = %v", err)
	}
	if n := s.count(message.RequestRetrieveInfo); n != 1 {
		t.Errorf("sent %d 0x7004 requests; a rejected credential must be tried exactly once", n)
	}
	if s.count(message.RequestEndServer) != 1 {
		t.Error("a clean rejection must still end the server job")
	}
}

func TestRejectionReasons(t *testing.T) {
	for code, reason := range map[uint32]signon.Reason{
		0x00020001: signon.ReasonProfileUnknown,
		0x00020002: signon.ReasonProfileDisabled,
		0x0003000C: signon.ReasonPasswordIncorrectNextDisables,
		0x0003000D: signon.ReasonPasswordExpired,
		0x00030010: signon.ReasonPasswordNone,
		0x00040000: signon.ReasonSecurityFailure,
		0x00050002: signon.ReasonExitProgram,
	} {
		s := &fakeSignon{passwordLevel: level(0), infoRC: code}
		_, err := run(t, s, testPassword, context.Background())
		var se *signon.Error
		if !errors.As(err, &se) || se.Reason() != reason || !errors.Is(err, signon.ErrAuthenticationRejected) {
			t.Errorf("0x%08X: error = %v; want %s", code, err, reason)
		}
		if s.count(message.RequestRetrieveInfo) != 1 {
			t.Errorf("0x%08X: retried", code)
		}
	}
}

func TestAttributeExchangeFailure(t *testing.T) {
	s := &fakeSignon{passwordLevel: level(0), attrRC: 0x00010004}
	_, err := run(t, s, testPassword, context.Background())
	var se *signon.Error
	if !errors.Is(err, signon.ErrRequestRejected) || errors.Is(err, signon.ErrAuthenticationRejected) || !errors.As(err, &se) || se.Stage != "exchange attributes" {
		t.Fatalf("error = %v", err)
	}
	if s.count(message.RequestRetrieveInfo) != 0 {
		t.Error("credential sent after a failed attribute exchange")
	}
}

func TestUnsupportedPasswordLevelSendsNoCredential(t *testing.T) {
	s := &fakeSignon{passwordLevel: level(4)}
	_, err := run(t, s, testPassword, context.Background())
	if !errors.Is(err, ibminet.ErrUnsupportedPasswordLevel) {
		t.Fatalf("error = %v", err)
	}
	if s.count(message.RequestRetrieveInfo) != 0 {
		t.Error("a credential was sent for password level 4")
	}
}

func TestMalformedRepliesCloseWithoutEndServer(t *testing.T) {
	for name, s := range map[string]*fakeSignon{
		"malformed 0xF003": {passwordLevel: level(0), malformedF003: true},
		"missing CCSID":    {passwordLevel: level(0), omitCCSID: true},
	} {
		_, err := run(t, s, testPassword, context.Background())
		if !errors.Is(err, hostserver.ErrMalformed) {
			t.Errorf("%s: error = %v", name, err)
		}
		if s.count(message.RequestEndServer) != 0 {
			t.Errorf("%s: sent 0x7006 after a malformed reply", name)
		}
	}
}

func TestWrongReplyIDPoisons(t *testing.T) {
	s := &fakeSignon{passwordLevel: level(0), wrongReplyID: true}
	_, err := run(t, s, testPassword, context.Background())
	if !errors.Is(err, hostserver.ErrUnexpectedReply) {
		t.Fatalf("error = %v", err)
	}
	if s.count(message.RequestRetrieveInfo) != 0 || s.count(message.RequestEndServer) != 0 {
		t.Errorf("requests after a poisoned stream: %04x", s.requests)
	}
}

func TestCancellationClosesSocket(t *testing.T) {
	s := &fakeSignon{silent: true}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := run(t, s, testPassword, ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v", err)
	}
}

func TestFreshSeedsPerValidation(t *testing.T) {
	var seeds [][8]byte
	for i := 0; i < 2; i++ {
		s := &fakeSignon{passwordLevel: level(0)}
		if _, err := run(t, s, testPassword, context.Background()); err != nil {
			t.Fatal(err)
		}
		seeds = append(seeds, s.clientSeeds[0])
		if !auth.ValidSeed(s.clientSeeds[0]) {
			t.Errorf("client seed %x outside the accepted range", s.clientSeeds[0])
		}
	}
	if seeds[0] == seeds[1] {
		t.Error("two validations reused a client seed")
	}
}

func TestErrorsAndInfoContainNoSecrets(t *testing.T) {
	s := &fakeSignon{passwordLevel: level(0)}
	_, err := run(t, s, "WRONGPW9", context.Background())
	ok := &fakeSignon{passwordLevel: level(2)}
	info, _ := run(t, ok, testPassword, context.Background())
	outputs := []string{err.Error(), fmt.Sprintf("%+v", err), fmt.Sprintf("%#v", info), fmt.Sprintf("%+v", *info)}
	secrets := []string{"CANARY", "WRONGPW9", testPassword, fmt.Sprintf("%x", serverSeed), fmt.Sprintf("%x", s.clientSeeds[0]), fmt.Sprintf("%x", s.substitutes[0]), fmt.Sprintf("%x", ok.substitutes[0])}
	for _, out := range outputs {
		for _, secret := range secrets {
			if strings.Contains(strings.ToUpper(out), strings.ToUpper(secret)) {
				t.Errorf("output leaks %q: %s", secret, out)
			}
		}
	}
}

func TestValidateRejectsWrongService(t *testing.T) {
	cred, _ := ibminet.NewPasswordCredential(testProfile, []byte(testPassword))
	ep := ibminet.Endpoint{Service: ibminet.ServiceRemoteCommand, Name: "as-rmtcmd", Host: "h", Port: 1, Source: ibminet.EndpointExplicit}
	if _, err := signon.Validate(context.Background(), ep, cred, signon.Options{}); err == nil {
		t.Error("Remote Command endpoint accepted")
	}
	if _, err := signon.Validate(context.Background(), ibminet.Endpoint{Service: ibminet.ServiceSignon}, nil, signon.Options{}); err == nil {
		t.Error("nil credential accepted")
	}
}

func TestReturnCodeReasons(t *testing.T) {
	for code, want := range map[signon.ReturnCode]signon.Reason{
		0x00010001: signon.ReasonRequestRejected,
		0x0001000B: signon.ReasonRequestRejected,
		0x00020003: signon.ReasonProfileMismatch,
		0x00020009: signon.ReasonProfileRejected,
		0x0003000B: signon.ReasonPasswordIncorrect,
		0x00030099: signon.ReasonPasswordRejected,
		0x00060001: signon.ReasonTokenFailure,
		0x00090000: signon.ReasonUnknown,
	} {
		if got := code.Reason(); got != want {
			t.Errorf("0x%08X = %s, want %s", uint32(code), got, want)
		}
	}
	unknown := &signon.Error{Stage: "x", Code: 0x00090000}
	if !errors.Is(unknown, signon.ErrRequestRejected) || errors.Is(unknown, signon.ErrAuthenticationRejected) {
		t.Error("unclassified codes must not be reported as credential rejections")
	}
}

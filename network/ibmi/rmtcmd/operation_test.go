package rmtcmd_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/hosttest"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd/message"
)

// session starts a session against srv and closes it at test end.
func session(t *testing.T, srv *hosttest.Server, opts rmtcmd.StartOptions) *rmtcmd.Session {
	t.Helper()
	s, err := startSession(t, srv, testPassword, ibminet.PasswordFamilyDES, opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close(time.Second) })
	return s
}

func mustCall(t *testing.T, params ...rpc.Parameter) rpc.Call {
	t.Helper()
	lib, _ := rpc.NewObjectName("TESTLIB")
	pgm, _ := rpc.NewObjectName("TESTPGM")
	c, err := rpc.NewCall(lib, pgm, params, rpc.MessagesUpToTen)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func must(p rpc.Parameter, err error) rpc.Parameter {
	if err != nil {
		panic(err)
	}
	return p
}

func TestRunCommandDSPJOB(t *testing.T) {
	var seen hostserver.Frame
	srv := &hosttest.Server{Handle: func(f hostserver.Frame) (hostserver.Frame, bool) {
		seen = f
		rest := append(hosttest.MessageRecord("CPC2191", 1, 0, "QCPFMSG", "QSYS", []byte{1, 2}, "Job displayed. "),
			hosttest.MessageRecord("CPI1234", 4, 10, "QCPFMSG", "QSYS", nil, "Info")...)
		return hosttest.OperationReply(message.ReplyRunCommand, f.Header.CorrelationID, 0, 2, rest), true
	}}
	s := session(t, srv, rmtcmd.StartOptions{})
	res, err := s.RunCommand(context.Background(), "DSPJOB", rmtcmd.RunCommandOptions{MessageOption: rpc.MessagesAll})
	if err != nil {
		t.Fatal(err)
	}
	// Level 7: option "all" is sent as 2 with CP 0x1101 and EBCDIC DSPJOB.
	want := []byte{0x02, 0, 0, 0, 12, 0x11, 0x01, 0xC4, 0xE2, 0xD7, 0xD1, 0xD6, 0xC2}
	if got := append(append([]byte(nil), seen.Template...), seen.Variable...); !bytes.Equal(got, want) {
		t.Errorf("request body = % x, want % x", got, want)
	}
	if res.ReturnCode != 0 || res.CorrelationID != 2 || res.EffectiveMessageOption != rpc.MessagesAll || len(res.Messages) != 2 {
		t.Fatalf("result = %+v", res)
	}
	m := res.Messages[0]
	if m.ID != "CPC2191" || m.Type != 1 || m.FileName != "QCPFMSG" || m.LibraryName != "QSYS" || m.Text != "Job displayed. " || !bytes.Equal(m.SubstitutionData, []byte{1, 2}) {
		t.Errorf("message = %+v; text must not be trimmed", m)
	}
	if res.Messages[1].Type != 4 || res.Messages[1].Severity != 10 {
		t.Errorf("second message = %+v", res.Messages[1])
	}
}

func TestRunCommandLayoutsByLevel(t *testing.T) {
	for _, tc := range []struct {
		level     uint16
		wantOpt   byte
		effective rpc.MessageOption
		wantCP    uint16
	}{
		{6, 0, rpc.MessagesUpToTen, 0x1101}, // "all" downgraded below level 7
		{7, 2, rpc.MessagesAll, 0x1101},
		{10, 4, rpc.MessagesAll, 0x1104},
		{11, 6, rpc.MessagesAll, 0x1104},
	} {
		var seen hostserver.Frame
		srv := &hosttest.Server{RemoteCommandLevel: tc.level, Handle: func(f hostserver.Frame) (hostserver.Frame, bool) {
			seen = f
			return hosttest.OperationReply(message.ReplyRunCommand, f.Header.CorrelationID, 0, 0, nil), true
		}}
		s := session(t, srv, rmtcmd.StartOptions{})
		res, err := s.RunCommand(context.Background(), "DSPJOB", rmtcmd.RunCommandOptions{MessageOption: rpc.MessagesAll})
		if err != nil {
			t.Fatal(err)
		}
		cp := binary.BigEndian.Uint16(seen.Variable[4:])
		if seen.Template[0] != tc.wantOpt || res.EffectiveMessageOption != tc.effective || cp != tc.wantCP {
			t.Errorf("level %d: option %d effective %s CP 0x%04X", tc.level, seen.Template[0], res.EffectiveMessageOption, cp)
		}
		// Level 10+ servers accept only UTF-16BE tagged 1200 (V7R5M0 rejects
		// EBCDIC tagged 37 with 0x0401); earlier levels take job-CCSID EBCDIC.
		want := []byte{0xC4, 0xE2, 0xD7, 0xD1, 0xD6, 0xC2} // DSPJOB, CCSID 37
		if tc.wantCP == 0x1104 {
			want = []byte{0, 0, 0x04, 0xB0, 0, 'D', 0, 'S', 0, 'P', 0, 'J', 0, 'O', 0, 'B'}
		}
		if got := seen.Variable[6:]; !bytes.Equal(got, want) {
			t.Errorf("level %d: command field % x, want % x", tc.level, got, want)
		}
	}
}

func TestRunCommandNativeFailure(t *testing.T) {
	for _, tc := range []struct {
		rc       uint16
		poisoned bool
	}{{0x0400, false}, {0x0303, false}, {0x0300, true}, {0x0305, true}, {0x0999, true}} {
		srv := &hosttest.Server{Handle: func(f hostserver.Frame) (hostserver.Frame, bool) {
			return hosttest.OperationReply(message.ReplyRunCommand, f.Header.CorrelationID, tc.rc, 1,
				hosttest.MessageRecord("CPF0001", 15, 40, "QCPFMSG", "QSYS", []byte("SECRETSUB"), "Error found on SECRETCMD")), true
		}}
		s := session(t, srv, rmtcmd.StartOptions{})
		res, err := s.RunCommand(context.Background(), "DLTF SECRETCMD", rmtcmd.RunCommandOptions{})
		var ne *rmtcmd.NativeError
		if !errors.Is(err, rmtcmd.ErrNative) || !errors.As(err, &ne) || ne.ReturnCode != tc.rc || res == nil || len(res.Messages) != 1 {
			t.Fatalf("0x%04X: result %+v, error %v", tc.rc, res, err)
		}
		if strings.Contains(err.Error(), "SECRET") || !strings.Contains(err.Error(), "CPF0001") {
			t.Errorf("0x%04X: error text %q", tc.rc, err)
		}
		if poisoned := errors.Is(s.Err(), hostserver.ErrPoisoned); poisoned != tc.poisoned {
			t.Errorf("0x%04X: poisoned = %t, want %t", tc.rc, poisoned, tc.poisoned)
		}
	}
}

func TestRunCommandRejectsBadInputBeforeSending(t *testing.T) {
	srv := &hosttest.Server{}
	s := session(t, srv, rmtcmd.StartOptions{})
	for name, cmd := range map[string]string{"empty": "", "NUL": "DSPJOB\x00", "unmappable": "SNDMSG MSG('€')"} {
		if _, err := s.RunCommand(context.Background(), cmd, rmtcmd.RunCommandOptions{}); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	if _, err := s.RunCommand(context.Background(), "DSPJOB", rmtcmd.RunCommandOptions{MessageOption: 9}); err == nil {
		t.Error("undefined message option accepted")
	}
	if srv.Count(message.RequestRunCommand) != 0 || s.Err() != nil {
		t.Error("invalid input reached the wire")
	}
}

// directionCall uses input, output, input/output and empty-output slots.
func directionCall(t *testing.T) rpc.Call {
	return mustCall(t,
		must(rpc.Input([]byte{1, 2, 0, 0})),
		must(rpc.Output(8)),
		must(rpc.InputOutput([]byte{0x41, 0}, 4)),
		must(rpc.Output(16)),
	)
}

func TestCallDirectionMatrix(t *testing.T) {
	for _, tc := range []struct {
		level      uint16
		inoutUsage uint16
	}{{4, 13}, {7, 33}} {
		var seen hostserver.Frame
		srv := &hosttest.Server{RemoteCommandLevel: tc.level, Handle: func(f hostserver.Frame) (hostserver.Frame, bool) {
			seen = f
			rest := append(hosttest.OutputEnvelope(12, 8, []byte{9, 9}), // zero-filled to 8
				hosttest.OutputEnvelope(tc.inoutUsage, 4, []byte{0x42, 0x43, 0x44, 0x45})...)
			rest = append(rest, hosttest.OutputEnvelope(12, 0, nil)...) // present, empty
			return hosttest.OperationReply(message.ReplyCallProgram, f.Header.CorrelationID, 0, 0, rest), true
		}}
		s := session(t, srv, rmtcmd.StartOptions{})
		call := directionCall(t)
		res, err := s.Call(context.Background(), call)
		if err != nil {
			t.Fatalf("level %d: %v", tc.level, err)
		}
		if err := res.Validate(call); err != nil {
			t.Fatalf("level %d: result misaligned: %v", tc.level, err)
		}
		// Request envelopes: usage 11 (trailing zeros dropped), 22, 13/33, 22.
		var usages []uint16
		for off := 0; off < len(seen.Variable); {
			p, n, err := message.ReadParameter(seen.Variable[off:])
			if err != nil {
				t.Fatal(err)
			}
			usages = append(usages, p.Usage)
			off += n
		}
		if want := []uint16{11, 22, tc.inoutUsage, 22}; len(usages) != 4 || usages[0] != want[0] || usages[1] != want[1] || usages[2] != want[2] || usages[3] != want[3] {
			t.Errorf("level %d: request usages %v, want %v", tc.level, usages, want)
		}
		if d, ok := res.Output(0); ok || d != nil {
			t.Error("input-only slot reported present")
		}
		if d, _ := res.Output(1); !bytes.Equal(d, []byte{9, 9, 0, 0, 0, 0, 0, 0}) {
			t.Errorf("output = % x; must be zero-filled to its length", d)
		}
		if d, _ := res.Output(2); !bytes.Equal(d, []byte{0x42, 0x43, 0x44, 0x45}) {
			t.Errorf("input/output = % x", d)
		}
		if d, ok := res.Output(3); !ok || d == nil || len(d) != 0 {
			t.Errorf("empty output = % x present %t", d, ok)
		}
	}
}

func TestCallCompressedOutput(t *testing.T) {
	srv := &hosttest.Server{Handle: func(f hostserver.Frame) (hostserver.Frame, bool) {
		rle := []byte{7, 0x1b, 0, 0, 0x00, 0x08, 0x1b, 0x1b} // 7, sixteen zeros, 0x1b
		return hosttest.OperationReply(message.ReplyCallProgram, f.Header.CorrelationID, 0, 0, hosttest.OutputEnvelope(22, 18, rle)), true
	}}
	s := session(t, srv, rmtcmd.StartOptions{})
	res, err := s.Call(context.Background(), mustCall(t, must(rpc.Output(64))))
	if err != nil {
		t.Fatal(err)
	}
	want := append(append([]byte{7}, make([]byte, 16)...), 0x1b)
	if d, _ := res.Output(0); !bytes.Equal(d, want) {
		t.Errorf("decompressed = % x", d)
	}
}

func TestCallNativeFailureKeepsMessages(t *testing.T) {
	for _, tc := range []struct {
		replyID  uint16
		rc       uint16
		poisoned bool
	}{{0x9003, 0x0500, false}, {0x8003, 0x0501, false}, {0x8003, 0x0400, false}, {0x9003, 0x0302, true}} {
		srv := &hosttest.Server{Handle: func(f hostserver.Frame) (hostserver.Frame, bool) {
			return hosttest.OperationReply(tc.replyID, f.Header.CorrelationID, tc.rc, 2, append(
				hosttest.MessageRecord("CPD0001", 2, 30, "QCPFMSG", "QSYS", nil, "diag"),
				hosttest.MessageRecord("CPF9801", 15, 40, "QCPFMSG", "QSYS", []byte("TESTPGM"), "Object not found")...)), true
		}}
		s := session(t, srv, rmtcmd.StartOptions{})
		call := mustCall(t, must(rpc.Output(8)))
		res, err := s.Call(context.Background(), call)
		var ne *rmtcmd.NativeError
		if !errors.As(err, &ne) || ne.ReturnCode != tc.rc || res == nil || len(res.Messages) != 2 || res.Messages[0].ID != "CPD0001" {
			t.Fatalf("0x%04X: result %+v, error %v", tc.rc, res, err)
		}
		if _, present := res.Output(0); present || res.Validate(call) != nil {
			t.Errorf("0x%04X: a failure body must not yield outputs", tc.rc)
		}
		if poisoned := errors.Is(s.Err(), hostserver.ErrPoisoned); poisoned != tc.poisoned {
			t.Errorf("0x%04X: poisoned = %t, want %t", tc.rc, poisoned, tc.poisoned)
		}
	}
}

func TestCallMalformedRepliesPoison(t *testing.T) {
	out := func(rest []byte) func(hostserver.Frame) (hostserver.Frame, bool) {
		return func(f hostserver.Frame) (hostserver.Frame, bool) {
			return hosttest.OperationReply(message.ReplyCallProgram, f.Header.CorrelationID, 0, 0, rest), true
		}
	}
	two := append(hosttest.OutputEnvelope(12, 8, nil), hosttest.OutputEnvelope(12, 8, nil)...)
	for name, handle := range map[string]func(hostserver.Frame) (hostserver.Frame, bool){
		"output over maximum": out(hosttest.OutputEnvelope(12, 9, nil)),
		"payload over length": out(hosttest.OutputEnvelope(12, 1, []byte{1, 2})),
		"bad usage":           out(hosttest.OutputEnvelope(11, 8, nil)),
		"too few outputs":     out(nil),
		"too many outputs":    out(two),
		"trailing bytes":      out(append(hosttest.OutputEnvelope(12, 8, nil), 0)),
		"RLE bomb":            out(hosttest.OutputEnvelope(22, 8, []byte{0x1b, 1, 1, 0xFF, 0xFF})),
		"RLE short":           out(hosttest.OutputEnvelope(22, 8, []byte{1})),
		"failure with rc zero": func(f hostserver.Frame) (hostserver.Frame, bool) {
			return hosttest.OperationReply(0x9003, f.Header.CorrelationID, 0, 0, nil), true
		},
		"bad message record": func(f hostserver.Frame) (hostserver.Frame, bool) {
			return hosttest.OperationReply(0x9003, f.Header.CorrelationID, 0x0500, 1, []byte{0, 0, 0, 5}), true
		},
		"wrong message count": func(f hostserver.Frame) (hostserver.Frame, bool) {
			return hosttest.OperationReply(0x9003, f.Header.CorrelationID, 0x0500, 2, hosttest.MessageRecord("CPF0001", 15, 40, "F", "L", nil, "x")), true
		},
	} {
		srv := &hosttest.Server{Handle: handle}
		s := session(t, srv, rmtcmd.StartOptions{})
		res, err := s.Call(context.Background(), mustCall(t, must(rpc.Output(8))))
		var oe *rmtcmd.OutcomeError
		if res != nil || !errors.As(err, &oe) || oe.Phase != rmtcmd.SendReplyReceived || !errors.Is(err, rmtcmd.ErrUncertainOutcome) {
			t.Errorf("%s: result %v, error %v", name, res, err)
			continue
		}
		if !errors.Is(s.Err(), hostserver.ErrPoisoned) {
			t.Errorf("%s: session not poisoned", name)
		}
	}
}

func TestCallReceiverLimitIsCheckedBeforeSending(t *testing.T) {
	srv := &hosttest.Server{}
	s := session(t, srv, rmtcmd.StartOptions{Limits: rmtcmd.Limits{MaxReceiver: 64}})
	_, err := s.Call(context.Background(), mustCall(t, must(rpc.Output(65))))
	if !errors.Is(err, rmtcmd.ErrLimit) {
		t.Fatalf("error = %v", err)
	}
	if srv.Count(message.RequestCallProgram) != 0 || s.Err() != nil {
		t.Error("an over-limit call reached the wire or poisoned the session")
	}
}

func TestCallNotStartedLeavesSessionUsable(t *testing.T) {
	srv := &hosttest.Server{Handle: func(f hostserver.Frame) (hostserver.Frame, bool) {
		return hosttest.OperationReply(message.ReplyCallProgram, f.Header.CorrelationID, 0, 0, hosttest.OutputEnvelope(12, 8, nil)), true
	}}
	s := session(t, srv, rmtcmd.StartOptions{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := s.Call(ctx, mustCall(t, must(rpc.Output(8))))
	var oe *rmtcmd.OutcomeError
	if !errors.As(err, &oe) || oe.Phase != rmtcmd.SendNotStarted || oe.MayHaveExecuted || errors.Is(err, rmtcmd.ErrUncertainOutcome) {
		t.Fatalf("error = %v", err)
	}
	if _, err := s.Call(context.Background(), mustCall(t, must(rpc.Output(8)))); err != nil {
		t.Fatalf("session unusable after a not-started call: %v", err)
	}
}

func TestCallLostReplyIsUncertainAndNotRetried(t *testing.T) {
	srv := &hosttest.Server{SilentOn: message.RequestCallProgram}
	s := session(t, srv, rmtcmd.StartOptions{})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := s.Call(ctx, mustCall(t, must(rpc.Output(8))))
	var oe *rmtcmd.OutcomeError
	if !errors.As(err, &oe) || oe.Phase != rmtcmd.SendFullFrameNoReply || !errors.Is(err, rmtcmd.ErrUncertainOutcome) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v", err)
	}
	if !errors.Is(s.Err(), hostserver.ErrPoisoned) {
		t.Error("session not poisoned after a lost reply")
	}
	if n := srv.Count(message.RequestCallProgram); n != 1 {
		t.Errorf("0x1003 sent %d times; a lost reply must never trigger a retry", n)
	}
}

// failingConn accepts only the first n bytes of the next write once armed.
type failingConn struct {
	net.Conn
	armed atomic.Bool
}

func (c *failingConn) Write(b []byte) (int, error) {
	if c.armed.Load() && len(b) > 10 {
		n, _ := c.Conn.Write(b[:10])
		return n, errors.New("injected write failure")
	}
	return c.Conn.Write(b)
}

func TestCallPartialWriteIsUncertain(t *testing.T) {
	srv := &hosttest.Server{Profile: testProfile, Password: testPassword}
	var fc *failingConn
	dial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		c, err := srv.Dial(ctx, network, addr)
		fc = &failingConn{Conn: c}
		return fc, err
	}
	conn, err := (&hostserver.Dialer{Dial: dial}).DialEndpoint(context.Background(), rmtEndpoint)
	if err != nil {
		t.Fatal(err)
	}
	cred, _ := ibminet.NewPasswordCredential(testProfile, []byte(testPassword))
	s, err := rmtcmd.Start(context.Background(), conn, cred, ibminet.PasswordFamilyDES, rmtcmd.StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close(time.Second)
	fc.armed.Store(true)
	_, err = s.Call(context.Background(), mustCall(t, must(rpc.Output(8))))
	var oe *rmtcmd.OutcomeError
	if !errors.As(err, &oe) || oe.Phase != rmtcmd.SendPartialFrame || !oe.MayHaveExecuted || !errors.Is(s.Err(), hostserver.ErrPoisoned) {
		t.Fatalf("error = %v, session %v", err, s.Err())
	}
}

func TestCorrelationsIncrease(t *testing.T) {
	var corrs []uint32
	srv := &hosttest.Server{Handle: func(f hostserver.Frame) (hostserver.Frame, bool) {
		corrs = append(corrs, f.Header.CorrelationID)
		return hosttest.OperationReply(message.ReplyRunCommand, f.Header.CorrelationID, 0, 0, nil), true
	}}
	s := session(t, srv, rmtcmd.StartOptions{})
	for i := 0; i < 3; i++ {
		if _, err := s.RunCommand(context.Background(), "DSPJOB", rmtcmd.RunCommandOptions{}); err != nil {
			t.Fatal(err)
		}
	}
	if len(corrs) != 3 || corrs[0] != 2 || corrs[1] != 3 || corrs[2] != 4 {
		t.Errorf("correlations = %v; 1 is the attribute exchange", corrs)
	}
}

// Level-11 servers answer message options 5 and 6 with 0x1107 (RCVM0300)
// records; Run Command and Program Call must decode them, help included.
func TestLevel11MessageRecords(t *testing.T) {
	rec := hosttest.RCVM0300Record("CPF9801", 15, 40, "QCPFMSG", "QSYS", []byte{1, 2}, "Object NOSUCHX in library QSYS not found.", "Cause . . . : The object was not found.")
	srv := &hosttest.Server{RemoteCommandLevel: 11, Handle: func(f hostserver.Frame) (hostserver.Frame, bool) {
		id := message.ReplyRunCommand
		if f.Header.RequestID != message.RequestRunCommand {
			id = 0x8003 // program call
		}
		return hosttest.OperationReply(id, f.Header.CorrelationID, 0x0400, 1, rec), true
	}}
	s := session(t, srv, rmtcmd.StartOptions{})
	res, err := s.RunCommand(context.Background(), "CHKOBJ OBJ(QSYS/NOSUCHX) OBJTYPE(*MSGQ)", rmtcmd.RunCommandOptions{MessageOption: rpc.MessagesAll})
	if !errors.Is(err, rmtcmd.ErrNative) || res == nil || len(res.Messages) != 1 {
		t.Fatalf("RunCommand = %+v, %v", res, err)
	}
	m := res.Messages[0]
	if m.ID != "CPF9801" || m.Type != 15 || m.Severity != 40 || m.FileName != "QCPFMSG" || m.LibraryName != "QSYS" ||
		m.Text != "Object NOSUCHX in library QSYS not found." || m.Help != "Cause . . . : The object was not found." || !bytes.Equal(m.SubstitutionData, []byte{1, 2}) {
		t.Errorf("message = %+v", m)
	}
	if err := s.Err(); err != nil {
		t.Fatalf("session poisoned: %v", err)
	}
}

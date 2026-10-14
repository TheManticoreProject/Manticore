package rmtcmd_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/hosttest"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd/message"
)

// TestSessionLostClassification checks which Call errors tell a neutral
// caller that the session is gone (so open handles must not be cleaned up
// with further calls).
func TestSessionLostClassification(t *testing.T) {
	failure := func(rc uint16) func(hostserver.Frame) (hostserver.Frame, bool) {
		return func(f hostserver.Frame) (hostserver.Frame, bool) {
			return hosttest.OperationReply(message.ReplyCallProgramFailure, f.Header.CorrelationID, rc, 1,
				hosttest.MessageRecord("CPF0001", 15, 40, "QCPFMSG", "QSYS", nil, "x")), true
		}
	}
	malformed := func(f hostserver.Frame) (hostserver.Frame, bool) {
		return hosttest.OperationReply(message.ReplyCallProgram, f.Header.CorrelationID, 0, 0, []byte{1}), true
	}
	for _, tc := range []struct {
		name   string
		handle func(hostserver.Frame) (hostserver.Frame, bool)
		silent bool
		cancel bool
		lost   bool
	}{
		{name: "known native failure", handle: failure(0x0500), lost: false},
		{name: "unknown native code", handle: failure(0x0302), lost: true},
		{name: "malformed reply", handle: malformed, lost: true},
		{name: "lost reply", silent: true, lost: true},
		{name: "cancelled before sending", handle: malformed, cancel: true, lost: false},
	} {
		srv := &hosttest.Server{Handle: tc.handle}
		if tc.silent {
			srv.SilentOn = message.RequestCallProgram
		}
		s := session(t, srv, rmtcmd.StartOptions{})
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		if tc.cancel {
			cancel()
		}
		_, err := s.Call(ctx, mustCall(t, must(rpc.Output(8))))
		cancel()
		if err == nil {
			t.Fatalf("%s: no error", tc.name)
		}
		if lost := errors.Is(err, rpc.ErrSessionLost); lost != tc.lost {
			t.Errorf("%s: ErrSessionLost = %t, want %t (%v)", tc.name, lost, tc.lost, err)
		}
		// Once lost, every later call also reports it without writing.
		if tc.lost {
			before := srv.Count(message.RequestCallProgram)
			if _, err := s.Call(context.Background(), mustCall(t, must(rpc.Output(8)))); !errors.Is(err, rpc.ErrSessionLost) {
				t.Errorf("%s: call after loss = %v", tc.name, err)
			}
			if srv.Count(message.RequestCallProgram) != before {
				t.Errorf("%s: a call on a lost session reached the wire", tc.name)
			}
		}
	}
}

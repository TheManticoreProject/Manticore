package client

import (
	"bytes"
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

func mustParam(p rpc.Parameter, err error) rpc.Parameter {
	if err != nil {
		panic(err)
	}
	return p
}

func testCall(t *testing.T, params ...rpc.Parameter) rpc.Call {
	t.Helper()
	lib, _ := rpc.NewObjectName("TESTLIB")
	pgm, _ := rpc.NewObjectName("TESTPGM")
	c, err := rpc.NewCall(lib, pgm, params, rpc.MessagesAll)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// operationServer answers 0x1002 with one completion message and 0x1003
// with outputs for (input, output 8, input/output 4).
func operationServer() *hosttest.Server {
	srv := newServer()
	srv.Handle = func(f hostserver.Frame) (hostserver.Frame, bool) {
		switch f.Header.RequestID {
		case message.RequestRunCommand:
			return hosttest.OperationReply(message.ReplyRunCommand, f.Header.CorrelationID, 0, 1,
				hosttest.MessageRecord("CPC2191", 1, 0, "QCPFMSG", "QSYS", nil, "Job displayed.")), true
		case message.RequestCallProgram:
			rest := append(hosttest.OutputEnvelope(12, 8, []byte("OUT")), hosttest.OutputEnvelope(33, 4, []byte("IO"))...)
			return hosttest.OperationReply(message.ReplyCallProgram, f.Header.CorrelationID, 0, 0, rest), true
		}
		return hostserver.Frame{}, false
	}
	return srv
}

// TestFacadeRunCommandAndCall is the Phase 5 exit gate: DSPJOB and a
// generic parameter-direction call complete through the public facade.
func TestFacadeRunCommandAndCall(t *testing.T) {
	srv := operationServer()
	c := newClient(t, srv)
	if err := c.Connect(context.Background(), credential(t, testPassword)); err != nil {
		t.Fatal(err)
	}
	res, err := c.RunCommand(context.Background(), "DSPJOB", rmtcmd.RunCommandOptions{})
	if err != nil || len(res.Messages) != 1 || res.Messages[0].ID != "CPC2191" {
		t.Fatalf("RunCommand = %+v, %v", res, err)
	}
	call := testCall(t, mustParam(rpc.Input([]byte{1})), mustParam(rpc.Output(8)), mustParam(rpc.InputOutput([]byte{2}, 4)))
	out, err := c.Call(context.Background(), call)
	if err != nil {
		t.Fatal(err)
	}
	if err := out.Validate(call); err != nil {
		t.Fatal(err)
	}
	if d, _ := out.Output(1); !bytes.Equal(d, []byte("OUT\x00\x00\x00\x00\x00")) {
		t.Errorf("output = %q", d)
	}
	if d, _ := out.Output(2); !bytes.Equal(d, []byte("IO\x00\x00")) {
		t.Errorf("input/output = %q", d)
	}
	if out.RequestedMessageOption != rpc.MessagesAll || out.EffectiveMessageOption != rpc.MessagesAll {
		t.Errorf("message options = %s / %s", out.RequestedMessageOption, out.EffectiveMessageOption)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
	if srv.Count(0x1004) != 1 {
		t.Error("Close did not end the server job")
	}
}

func TestNativeFailureKeepsClientReady(t *testing.T) {
	srv := newServer()
	srv.Handle = func(f hostserver.Frame) (hostserver.Frame, bool) {
		return hosttest.OperationReply(message.ReplyCallProgramFailure, f.Header.CorrelationID, 0x0500, 1,
			hosttest.MessageRecord("CPF9801", 15, 40, "QCPFMSG", "QSYS", nil, "Object not found.")), true
	}
	c := newClient(t, srv)
	if err := c.Connect(context.Background(), credential(t, testPassword)); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	res, err := c.Call(context.Background(), testCall(t, mustParam(rpc.Output(8))))
	if !errors.Is(err, rmtcmd.ErrNative) || res == nil || len(res.Messages) != 1 {
		t.Fatalf("Call = %+v, %v", res, err)
	}
	if c.getState() != stateReady {
		t.Errorf("state = %s; a known server failure keeps the session", c.getState())
	}
}

func TestMalformedReplyPoisonsClient(t *testing.T) {
	srv := newServer()
	srv.Handle = func(f hostserver.Frame) (hostserver.Frame, bool) {
		return hosttest.OperationReply(message.ReplyCallProgram, f.Header.CorrelationID, 0, 0, []byte{0, 0, 0, 1}), true
	}
	c := newClient(t, srv)
	if err := c.Connect(context.Background(), credential(t, testPassword)); err != nil {
		t.Fatal(err)
	}
	_, err := c.Call(context.Background(), testCall(t, mustParam(rpc.Output(8))))
	if !errors.Is(err, rmtcmd.ErrUncertainOutcome) {
		t.Fatalf("Call = %v", err)
	}
	if c.getState() != statePoisoned || c.SessionInfo().Connected {
		t.Fatalf("state = %s", c.getState())
	}
	if _, err := c.Call(context.Background(), testCall(t)); !errors.Is(err, ErrNotReady) || !errors.Is(err, rpc.ErrSessionLost) {
		t.Errorf("Call on a poisoned client = %v", err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
	if srv.Count(0x1004) != 0 {
		t.Error("0x1004 written to a poisoned stream")
	}
}

func TestExecuteExclusive(t *testing.T) {
	srv := operationServer()
	c := newClient(t, srv)
	if err := c.Connect(context.Background(), credential(t, testPassword)); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	call := testCall(t, mustParam(rpc.Input([]byte{1})), mustParam(rpc.Output(8)), mustParam(rpc.InputOutput([]byte{2}, 4)))
	var kept rpc.Caller
	err := c.ExecuteExclusive(context.Background(), func(caller rpc.Caller) error {
		kept = caller
		for i := 0; i < 3; i++ {
			if _, err := caller.Call(context.Background(), call); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if srv.Count(message.RequestCallProgram) != 3 {
		t.Errorf("calls = %d", srv.Count(message.RequestCallProgram))
	}
	if _, err := kept.Call(context.Background(), call); !errors.Is(err, ErrReservationReleased) {
		t.Errorf("caller used after return = %v", err)
	}
	if srv.Count(message.RequestCallProgram) != 3 {
		t.Error("a released caller reached the wire")
	}
	sentinel := errors.New("callback failure")
	if err := c.ExecuteExclusive(context.Background(), func(rpc.Caller) error { return sentinel }); !errors.Is(err, sentinel) {
		t.Errorf("callback error = %v", err)
	}
}

func TestLimitConfig(t *testing.T) {
	c, err := New(Config{Host: "h", Transport: 2})
	if err != nil {
		t.Fatal(err)
	}
	if c.cfg.MaxReceiver != rmtcmd.DefaultMaxReceiver || c.cfg.MaxMessageText != rmtcmd.DefaultMaxMessageText ||
		c.cfg.MaxRLEExpansion != rmtcmd.DefaultMaxRLEExpansion || c.cfg.MaxMessages != rmtcmd.DefaultMaxMessages {
		t.Errorf("limit defaults = %+v", c.cfg)
	}
	for name, mutate := range map[string]func(*Config){
		"receiver":     func(c *Config) { c.MaxReceiver = rmtcmd.MaxLimit + 1 },
		"RLE":          func(c *Config) { c.MaxRLEExpansion = rmtcmd.MaxLimit + 1 },
		"messages":     func(c *Config) { c.MaxMessages = 70000 },
		"text > frame": func(c *Config) { c.MaxFrame, c.MaxMessageText = 1024, 2048 },
	} {
		cfg := Config{Host: "h", Transport: 2}
		mutate(&cfg)
		if _, err := New(cfg); !errors.Is(err, ErrConfig) {
			t.Errorf("%s: New = %v", name, err)
		}
	}
}

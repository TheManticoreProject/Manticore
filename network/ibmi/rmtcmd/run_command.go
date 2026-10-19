package rmtcmd

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd/message"
)

// RunCommandOptions configures RunCommand.
type RunCommandOptions struct {
	// MessageOption selects returned messages; zero means MessagesUpToTen.
	MessageOption rpc.MessageOption
}

// CommandResult is the structured reply to a CL command: the Remote Command
// return code and ordered IBM messages. There is no stdout stream.
type CommandResult struct {
	ReturnCode             uint16
	Messages               []rpc.IBMMessage
	RequestedMessageOption rpc.MessageOption
	EffectiveMessageOption rpc.MessageOption
	CorrelationID          uint32
}

// RunCommand executes raw IBM i CL source (request 0x1002).
//
// This is a privileged operation: command is passed to the server verbatim,
// so any caller-controlled text inside it is CL injection. Prefer a typed
// program-call API when one exists; never build commands by concatenating
// untrusted input.
//
// A nonzero return code returns the CommandResult together with a
// *NativeError. Transport and protocol failures return an *OutcomeError;
// a command that may have run is never retried. Neither error text includes
// the command or message text.
func (s *Session) RunCommand(ctx context.Context, command string, opts RunCommandOptions) (*CommandResult, error) {
	const op = "run command"
	if command == "" {
		return nil, fmt.Errorf("rmtcmd: %s: empty command", op)
	}
	if strings.IndexByte(command, 0) >= 0 {
		return nil, fmt.Errorf("rmtcmd: %s: command contains NUL", op)
	}
	requested := opts.MessageOption
	if requested == 0 {
		requested = rpc.MessagesUpToTen
	}
	if !requested.Valid() {
		return nil, fmt.Errorf("rmtcmd: %s: undefined message option %s", op, requested)
	}
	var req message.RunCommandRequest
	if s.info.Capabilities.ExplicitCommandCCSID {
		// Level 10+ servers accept the CCSID-tagged command only as
		// UTF-16BE tagged 1200; EBCDIC tagged with the job CCSID fails
		// with 0x0401 (observed on V7R5M0).
		if !utf8.ValidString(command) {
			return nil, fmt.Errorf("rmtcmd: %s: command is not valid UTF-8", op)
		}
		ccsid := message.CommandCCSIDUTF16
		req.Command, req.CommandCCSID = utf16BE(command), &ccsid
	} else {
		encoded, err := s.codec.Encode(command)
		if err != nil {
			return nil, fmt.Errorf("rmtcmd: %s: %w", op, err)
		}
		req.Command = encoded
	}
	var effective rpc.MessageOption
	req.MessageOption, effective = wireMessageOption(requested, s.info.Capabilities)
	var err error
	if req.Correlation, err = s.corr.Next(); err != nil {
		return nil, err
	}
	frame, err := req.Frame()
	if err != nil {
		return nil, fmt.Errorf("rmtcmd: %s: %w", op, err)
	}
	size, err := frame.Size()
	if err != nil {
		return nil, fmt.Errorf("rmtcmd: %s: %w", op, err)
	}

	reply, receipt, err := s.conn.Exchange(ctx, frame, hostserver.Expect{
		ServerID: hostserver.ServerIDRemoteCommand, ReplyIDs: []uint16{message.ReplyRunCommand}, CorrelationID: req.Correlation,
	})
	clear(frame.Variable) // the command may contain secrets
	if err != nil {
		return nil, exchangeOutcome(op, receipt, size, err)
	}
	r, err := message.ParseOperationReply(reply)
	if err != nil {
		return nil, replyOutcome(op, s.conn.Poison(err))
	}
	// Run Command replies carry messages whatever the return code.
	msgs, err := s.decodeMessages(r.MessageCount, r.Rest)
	if err != nil {
		return nil, replyOutcome(op, s.conn.Poison(err))
	}
	result := &CommandResult{
		ReturnCode:             r.ReturnCode,
		Messages:               msgs,
		RequestedMessageOption: requested,
		EffectiveMessageOption: effective,
		CorrelationID:          req.Correlation,
	}
	if r.ReturnCode == 0 {
		return result, nil
	}
	native := &NativeError{Operation: op, ReturnCode: r.ReturnCode, Messages: msgs}
	if !keepsSession(r.ReturnCode, true) {
		native.SessionLost = true
		_ = s.conn.Poison(native)
	}
	return result, native
}

// utf16BE encodes s as big-endian UTF-16 code units, with no BOM.
func utf16BE(s string) []byte {
	units := utf16.Encode([]rune(s))
	b := make([]byte, 0, 2*len(units))
	for _, u := range units {
		b = append(b, byte(u>>8), byte(u))
	}
	return b
}

package rmtcmd

import (
	"context"
	"fmt"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd/message"
)

var _ rpc.Caller = (*Session)(nil)

// Call runs one IBM i program call (request 0x1003). It implements
// rpc.Caller.
//
// On success the Result aligns one slot per parameter, with output-capable
// slots present. When IBM i reports failure (0x9003, or a nonzero return
// code) Call returns a Result carrying the messages, all slots absent,
// together with a *NativeError. Transport and protocol failures return an
// *OutcomeError recording whether the program may have run; the call is
// never retried.
//
// Session methods are not safe for concurrent use; network/ibmi/client
// serializes them.
func (s *Session) Call(ctx context.Context, call rpc.Call) (*rpc.Result, error) {
	if call.IsZero() {
		return nil, fmt.Errorf("rmtcmd: %w", &rpc.ValidationError{Field: "call", Reason: "zero Call"})
	}
	op := "call program " + call.String()
	program, err := ccsid.EncodeFixed(s.codec, call.Program().String(), 10)
	if err != nil {
		return nil, fmt.Errorf("rmtcmd: %s: program name: %w", op, err)
	}
	library, err := ccsid.EncodeFixed(s.codec, call.Library().String(), 10)
	if err != nil {
		return nil, fmt.Errorf("rmtcmd: %s: library name: %w", op, err)
	}
	params := call.Parameters()
	req := message.CallProgramRequest{Parameters: make([]message.RequestParameter, len(params))}
	copy(req.Program[:], program)
	copy(req.Library[:], library)
	for i, p := range params {
		if p.Direction() != rpc.DirectionInput && p.Maximum() > s.limits.MaxReceiver {
			return nil, fmt.Errorf("%w: %s: parameter %d receiver of %d bytes exceeds %d", ErrLimit, op, i, p.Maximum(), s.limits.MaxReceiver)
		}
		rp := message.RequestParameter{Maximum: p.Maximum(), Data: message.TrimTrailingZeros(p.InputBytes())}
		switch p.Direction() {
		case rpc.DirectionInput:
			rp.Usage = message.UsageInput
		case rpc.DirectionOutput:
			rp.Usage = message.UsageOutput
		case rpc.DirectionInputOutput:
			rp.Usage = message.UsageInputOutput
			if s.info.Capabilities.InputOutputUsage33 {
				rp.Usage = message.UsageInputOutputNew
			}
		}
		req.Parameters[i] = rp
	}
	var effective rpc.MessageOption
	req.MessageOption, effective = wireMessageOption(call.MessageOption(), s.info.Capabilities)
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
		ServerID:      hostserver.ServerIDRemoteCommand,
		ReplyIDs:      []uint16{message.ReplyCallProgram, message.ReplyCallProgramFailure},
		CorrelationID: req.Correlation,
	})
	clear(frame.Variable) // input parameters may contain secrets
	if err != nil {
		return nil, exchangeOutcome(op, receipt, size, err)
	}
	r, err := message.ParseOperationReply(reply)
	if err != nil {
		return nil, replyOutcome(op, s.conn.Poison(err))
	}

	result := &rpc.Result{
		Parameters:             make([]rpc.ParameterResult, len(params)),
		RequestedMessageOption: call.MessageOption(),
		EffectiveMessageOption: effective,
		CorrelationID:          req.Correlation,
	}
	if r.ReturnCode == 0 {
		if r.ReplyID == message.ReplyCallProgramFailure {
			return nil, replyOutcome(op, s.conn.Poison(&hostserver.MalformedError{What: "reply", Offset: 20, Reason: "failure reply 0x9003 with return code 0"}))
		}
		if err := s.decodeOutputs(params, r.Rest, result.Parameters); err != nil {
			return nil, replyOutcome(op, s.conn.Poison(err))
		}
		return result, nil
	}

	// A nonzero return code selects a message body; outputs are not decoded.
	if result.Messages, err = s.decodeMessages(r.MessageCount, r.Rest); err != nil {
		return nil, replyOutcome(op, s.conn.Poison(err))
	}
	native := &NativeError{Operation: op, ReturnCode: r.ReturnCode, Messages: result.Messages}
	if !keepsSession(r.ReturnCode, false) {
		_ = s.conn.Poison(native)
	}
	return result, native
}

// replyUsageOK reports whether a returned parameter's usage is valid for its
// request direction at the negotiated level, and whether it is compressed.
func (s *Session) replyUsageOK(d rpc.Direction, usage uint16) (ok, compressed bool) {
	switch d {
	case rpc.DirectionOutput:
		switch usage {
		case message.ReplyUsageOutputBase, message.ReplyUsageOutput:
			return true, false
		case message.ReplyUsageOutputCompressed:
			return true, true
		}
	case rpc.DirectionInputOutput:
		switch usage {
		case message.ReplyUsageInputOutputBase:
			return true, false
		case message.ReplyUsageInputOutputCompressed:
			return true, true
		case message.UsageInputOutput:
			return !s.info.Capabilities.InputOutputUsage33, false
		case message.UsageInputOutputNew:
			return s.info.Capabilities.InputOutputUsage33, false
		}
	}
	return false, false
}

// decodeOutputs reads one envelope per output-capable parameter, in request
// order, and requires b to be consumed exactly.
func (s *Session) decodeOutputs(params []rpc.Parameter, b []byte, slots []rpc.ParameterResult) error {
	off := 0
	for i, p := range params {
		if p.Direction() == rpc.DirectionInput {
			continue
		}
		rp, n, err := message.ReadParameter(b[off:])
		if err != nil {
			return fmt.Errorf("parameter %d: %w", i, err)
		}
		ok, compressed := s.replyUsageOK(p.Direction(), rp.Usage)
		if !ok {
			return &hostserver.MalformedError{What: "parameter", Offset: 24 + off + 10, Reason: fmt.Sprintf("parameter %d has reply usage %d", i, rp.Usage)}
		}
		if rp.Length > p.Maximum() || rp.Length > s.limits.MaxReceiver {
			return &hostserver.MalformedError{What: "parameter", Offset: 24 + off + 6, Reason: fmt.Sprintf("parameter %d returns %d bytes, maximum %d", i, rp.Length, min(p.Maximum(), s.limits.MaxReceiver))}
		}
		var data []byte
		if compressed {
			if rp.Length > s.limits.MaxRLEExpansion {
				return fmt.Errorf("%w: parameter %d expands to %d bytes, limit %d", ErrLimit, i, rp.Length, s.limits.MaxRLEExpansion)
			}
			if data, err = message.DecodeRLE(rp.Payload, rp.Length); err != nil {
				return fmt.Errorf("parameter %d: %w", i, err)
			}
		} else {
			if uint64(len(rp.Payload)) > uint64(rp.Length) {
				return &hostserver.MalformedError{What: "parameter", Offset: 24 + off, Reason: fmt.Sprintf("parameter %d carries %d bytes for a %d-byte output", i, len(rp.Payload), rp.Length)}
			}
			data = make([]byte, rp.Length) // the server may omit trailing zeros
			copy(data, rp.Payload)
		}
		slots[i] = rpc.ParameterResult{Present: true, Data: data}
		off += n
	}
	if off != len(b) {
		return &hostserver.MalformedError{What: "reply", Offset: 24 + off, Reason: fmt.Sprintf("%d unexplained bytes after the output parameters", len(b)-off)}
	}
	return nil
}

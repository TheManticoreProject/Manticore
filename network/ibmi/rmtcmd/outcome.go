package rmtcmd

import (
	"errors"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// SendPhase records how far a request got before an operation failed.
type SendPhase uint8

const (
	// SendNotStarted: no request byte was accepted; nothing ran.
	SendNotStarted SendPhase = iota
	// SendPartialFrame: part of the request was accepted; the stream is
	// poisoned and the outcome is unknown.
	SendPartialFrame
	// SendFullFrameNoReply: the whole request was accepted locally but no
	// valid reply arrived; the server may have executed it.
	SendFullFrameNoReply
	// SendReplyReceived: a reply arrived but could not be used; the server
	// executed the request.
	SendReplyReceived
)

func (p SendPhase) String() string {
	switch p {
	case SendNotStarted:
		return "not started"
	case SendPartialFrame:
		return "partial frame sent"
	case SendFullFrameNoReply:
		return "full frame sent, no reply"
	case SendReplyReceived:
		return "reply received"
	}
	return fmt.Sprintf("SendPhase(%d)", uint8(p))
}

// ErrUncertainOutcome is matched by an OutcomeError whose request may have
// executed. The library never retries or replays such a request.
var ErrUncertainOutcome = errors.New("rmtcmd: operation outcome unknown")

// OutcomeError reports a transport or protocol failure of an operation,
// with how far the request got. "Accepted" means accepted by the local
// connection, not received or executed by the server.
type OutcomeError struct {
	Operation       string
	Phase           SendPhase
	MayHaveExecuted bool
	Err             error
}

func (e *OutcomeError) Error() string {
	if e.MayHaveExecuted {
		return fmt.Sprintf("rmtcmd: %s failed (%s; it may have executed): %v", e.Operation, e.Phase, e.Err)
	}
	return fmt.Sprintf("rmtcmd: %s failed (%s): %v", e.Operation, e.Phase, e.Err)
}

func (e *OutcomeError) Unwrap() error { return e.Err }

// Is makes errors.Is(err, ErrUncertainOutcome) true when the request may
// have executed.
func (e *OutcomeError) Is(target error) bool {
	return target == ErrUncertainOutcome && e.MayHaveExecuted
}

// exchangeOutcome classifies an Exchange failure for a frame of size bytes.
func exchangeOutcome(op string, receipt hostserver.SendReceipt, size uint32, err error) *OutcomeError {
	phase := SendNotStarted
	switch {
	case receipt.BytesAccepted == 0:
	case uint64(receipt.BytesAccepted) < uint64(size):
		phase = SendPartialFrame
	default:
		phase = SendFullFrameNoReply
	}
	return &OutcomeError{Operation: op, Phase: phase, MayHaveExecuted: phase != SendNotStarted, Err: err}
}

// replyOutcome reports an unusable reply; the server executed the request.
func replyOutcome(op string, err error) *OutcomeError {
	return &OutcomeError{Operation: op, Phase: SendReplyReceived, MayHaveExecuted: true, Err: err}
}

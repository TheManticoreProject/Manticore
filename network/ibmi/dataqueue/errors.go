package dataqueue

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/TheManticoreProject/Manticore/network/ibmi/dataqueue/message"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// Sentinels matched (via errors.Is) by an *Error, from the IBM message ID
// of a command-check reply (return code 0xF001).
var (
	ErrQueueNotFound     = errors.New("dataqueue: queue not found")
	ErrLibraryNotFound   = errors.New("dataqueue: library not found")
	ErrNotAuthorized     = errors.New("dataqueue: not authorized")
	ErrAlreadyExists     = errors.New("dataqueue: queue already exists")
	ErrWrongQueueKind    = errors.New("dataqueue: operation does not match the queue kind")
	ErrRejectedByExit    = errors.New("dataqueue: rejected by an exit program")
	ErrServerRejected    = errors.New("dataqueue: request rejected by the server")
	ErrUncertainOutcome  = errors.New("dataqueue: operation outcome unknown")
	ErrUnsupportedLength = errors.New("dataqueue: unsupported queue or entry length")
)

var messageSentinels = map[string]error{
	"CPF9810": ErrLibraryNotFound,
	"CPF9801": ErrQueueNotFound,
	"CPF2105": ErrQueueNotFound,
	"CPF9802": ErrNotAuthorized,
	"CPF2189": ErrNotAuthorized,
	"CPF9820": ErrNotAuthorized,
	"CPF2182": ErrNotAuthorized,
	"CPF9870": ErrAlreadyExists,
	"CPF9502": ErrWrongQueueKind,
	"CPF9506": ErrWrongQueueKind,
}

var returnCodeText = map[uint16]string{
	message.RCCommandCheck:         "command check",
	message.RCProtocolError:        "protocol error",
	message.RCSyntaxError:          "syntax error",
	message.RCQueueDestroyed:       "queue destroyed",
	message.RCUnsupportedLength:    "unsupported queue or entry length",
	message.RCInvalidLevel:         "invalid data-stream level",
	message.RCInvalidVRM:           "invalid version, release or modification",
	message.RCExitRejected:         "rejected by an exit program",
	message.RCExitNotAuthorized:    "exit program not authorized",
	message.RCExitNotFound:         "exit program not found",
	message.RCExitFailed:           "exit program failed",
	message.RCExitInvalidProgramNo: "invalid number of exit programs",
}

// Error is a server-reported failure (common reply 0x8002). The session
// remains usable.
type Error struct {
	Operation  string
	ReturnCode uint16
	// MessageID is the IBM message ID of a command check, for example
	// "CPF9801", or "".
	MessageID string
	// Text is the decoded message, which may quote object names. It is
	// never part of Error().
	Text string
}

func (e *Error) Error() string {
	what, ok := returnCodeText[e.ReturnCode]
	if !ok {
		what = "unknown return code"
	}
	if e.MessageID != "" {
		return fmt.Sprintf("dataqueue: %s failed: return code 0x%04X (%s, %s)", e.Operation, e.ReturnCode, what, e.MessageID)
	}
	return fmt.Sprintf("dataqueue: %s failed: return code 0x%04X (%s)", e.Operation, e.ReturnCode, what)
}

// Is matches ErrServerRejected, the message-ID sentinels, ErrRejectedByExit
// for the exit-program codes and ErrUnsupportedLength.
func (e *Error) Is(target error) bool {
	switch target {
	case ErrServerRejected:
		return true
	case ErrRejectedByExit:
		return e.ReturnCode >= message.RCExitRejected && e.ReturnCode <= message.RCExitInvalidProgramNo
	case ErrUnsupportedLength:
		return e.ReturnCode == message.RCUnsupportedLength
	}
	return target != nil && messageSentinels[e.MessageID] == target
}

var messageID = regexp.MustCompile(`^[A-Z][A-Z0-9]{2}[0-9A-F]{4}`)

// OutcomeError reports a transport or protocol failure. The connection is
// poisoned and the Session unusable. MayHaveExecuted is true once any
// request byte was accepted locally: the server may have run it.
type OutcomeError struct {
	Operation       string
	MayHaveExecuted bool
	Err             error
}

func (e *OutcomeError) Error() string {
	if e.MayHaveExecuted {
		return fmt.Sprintf("dataqueue: %s failed (it may have executed): %v", e.Operation, e.Err)
	}
	return fmt.Sprintf("dataqueue: %s failed: %v", e.Operation, e.Err)
}

func (e *OutcomeError) Unwrap() error { return e.Err }

// Is makes errors.Is(err, ErrUncertainOutcome) true when the request may
// have executed.
func (e *OutcomeError) Is(target error) bool {
	return target == ErrUncertainOutcome && e.MayHaveExecuted
}

func exchangeOutcome(op string, receipt hostserver.SendReceipt, err error) *OutcomeError {
	return &OutcomeError{Operation: op, MayHaveExecuted: receipt.BytesAccepted > 0, Err: err}
}

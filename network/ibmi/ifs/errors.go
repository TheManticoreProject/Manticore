package ifs

import (
	"errors"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// Sentinels matched (via errors.Is) by an *Error.
var (
	ErrServerRejected = errors.New("ifs: request rejected by the server")
	ErrNotFound       = errors.New("ifs: file or path not found")
	ErrAccessDenied   = errors.New("ifs: access denied")
	ErrInUse          = errors.New("ifs: file in use or locked")
)

// Other errors.
var (
	// ErrUncertainOutcome is matched by an *OutcomeError whose request
	// may have reached the server.
	ErrUncertainOutcome = errors.New("ifs: operation outcome unknown")
	// ErrFileClosed is returned by operations on a closed File.
	ErrFileClosed = errors.New("ifs: file already closed")
	// ErrTooLarge is returned by ReadFile when the file exceeds the limit.
	ErrTooLarge = errors.New("ifs: file larger than the limit")
	// ErrUnsupported is matched when the server negotiates a data-stream
	// level or pathname CCSID this package does not implement.
	ErrUnsupported = errors.New("ifs: unsupported server attributes")
)

// Return codes of the generic reply 0x8001 (docs/08 section 9).
const (
	RCSuccess          uint16 = 0
	RCFileInUse        uint16 = 1
	RCFileNotFound     uint16 = 2
	RCPathNotFound     uint16 = 3
	RCAccessDenied     uint16 = 5
	RCInvalidHandle    uint16 = 6
	RCRequestDenied    uint16 = 13
	RCNoMoreData       uint16 = 22
	RCSharingViolation uint16 = 32
	RCLockViolation    uint16 = 33
)

var returnCodeText = map[uint16]string{
	1: "file in use", 2: "file not found", 3: "path not found", 4: "duplicate directory-entry name",
	5: "access denied to directory entry", 6: "invalid handle", 7: "invalid directory-entry name",
	8: "invalid attribute name", 9: "directory not empty", 10: "file substream in use",
	11: "resource limit exceeded", 12: "resource not available", 13: "access denied to request",
	14: "directory entry damaged", 15: "invalid connection", 16: "invalid request",
	17: "data-stream syntax error", 18: "no more files", 19: "parameter not supported",
	20: "parameter value not supported", 21: "cannot convert value", 22: "no more data",
	23: "request not supported", 24: "invalid user", 25: "unknown error",
	32: "sharing violation", 33: "lock violation", 34: "stale handle",
	0x0D42: "invalid file handle", 0x125E: "closed with damage",
}

// Error is a return code reported by the server. The session remains
// usable.
type Error struct {
	Operation  string
	ReturnCode uint16
}

func (e *Error) Error() string {
	what, ok := returnCodeText[e.ReturnCode]
	if !ok {
		what = "unknown return code"
	}
	return fmt.Sprintf("ifs: %s failed: return code %d (%s)", e.Operation, e.ReturnCode, what)
}

// Is matches ErrServerRejected and the category sentinels.
func (e *Error) Is(target error) bool {
	switch target {
	case ErrServerRejected:
		return true
	case ErrNotFound:
		return e.ReturnCode == RCFileNotFound || e.ReturnCode == RCPathNotFound
	case ErrAccessDenied:
		return e.ReturnCode == RCAccessDenied || e.ReturnCode == RCRequestDenied
	case ErrInUse:
		return e.ReturnCode == RCFileInUse || e.ReturnCode == 10 || e.ReturnCode == RCSharingViolation || e.ReturnCode == RCLockViolation
	}
	return false
}

// OutcomeError reports a transport or protocol failure. The connection is
// poisoned and the Session unusable. MayHaveExecuted is true once any
// request byte was accepted locally.
type OutcomeError struct {
	Operation       string
	MayHaveExecuted bool
	Err             error
}

func (e *OutcomeError) Error() string {
	if e.MayHaveExecuted {
		return fmt.Sprintf("ifs: %s failed (it may have executed): %v", e.Operation, e.Err)
	}
	return fmt.Sprintf("ifs: %s failed: %v", e.Operation, e.Err)
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

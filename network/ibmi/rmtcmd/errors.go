package rmtcmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
)

// ErrAttributesRejected is matched when the 0x8001 return code is not a
// success or a documented warning.
var ErrAttributesRejected = errors.New("rmtcmd: exchange attributes rejected")

// AttributesError reports a fatal 0x8001 return code.
type AttributesError struct {
	Code uint16
}

func (e *AttributesError) Error() string {
	return fmt.Sprintf("rmtcmd: exchange attributes failed with return code 0x%04X (%s)", e.Code, attributesCodeText(e.Code))
}

// Is makes errors.Is(err, ErrAttributesRejected) true.
func (e *AttributesError) Is(target error) bool { return target == ErrAttributesRejected }

// attributesWarnings are the 0x8001 codes after which the session remains
// usable (iSeriesNavigatorReverse docs/19, "Exchange attributes"). 0x0104
// (invalid CCSID) is accepted only because the job CCSID is validated
// separately before the session becomes ready.
var attributesWarnings = map[uint16]string{
	0x0100: "limited user",
	0x0104: "invalid CCSID",
	0x0105: "invalid NLV; server uses its primary NLV",
	0x0106: "NLV not installed; server uses its primary NLV",
	0x0107: "product information unavailable for NLV validation",
	0x0108: "NLV library could not be added",
}

var attributesErrors = map[uint16]string{
	0x0101: "invalid exchange request",
	0x0102: "invalid data-stream level",
	0x0103: "invalid client version",
	0x0200: "request processing error",
	0x0201: "invalid LL",
	0x0202: "invalid server ID",
	0x0203: "incomplete data",
	0x0204: "host resource error",
	0x0205: "invalid request ID",
}

func attributesCodeText(code uint16) string {
	if t, ok := attributesWarnings[code]; ok {
		return t
	}
	if t, ok := attributesErrors[code]; ok {
		return t
	}
	return "unknown"
}

// ErrNative is matched by a NativeError: IBM i completed the request and
// reported failure with a Remote Command return code.
var ErrNative = errors.New("rmtcmd: operation failed on the server")

// NativeError is a nonzero Remote Command return code with the IBM messages
// that explain it. Its text lists at most five message IDs and never the
// message text, substitution data or command source, which can contain
// secrets.
type NativeError struct {
	Operation  string
	ReturnCode uint16
	Messages   []rpc.IBMMessage
	// SessionLost is set when the return code's body semantics are unknown,
	// so the session was poisoned.
	SessionLost bool
}

func (e *NativeError) Error() string {
	ids := make([]string, 0, 5)
	for _, m := range e.Messages {
		if m.ID != "" && len(ids) < 5 {
			ids = append(ids, m.ID)
		}
	}
	if len(ids) == 0 {
		return fmt.Sprintf("rmtcmd: %s failed with return code 0x%04X", e.Operation, e.ReturnCode)
	}
	return fmt.Sprintf("rmtcmd: %s failed with return code 0x%04X (%s)", e.Operation, e.ReturnCode, strings.Join(ids, ", "))
}

// Is makes errors.Is(err, ErrNative) true, and errors.Is(err,
// rpc.ErrSessionLost) true when the session was poisoned.
func (e *NativeError) Is(target error) bool {
	return target == ErrNative || (target == rpc.ErrSessionLost && e.SessionLost)
}

// ErrLimit is matched when a request or reply exceeds a configured limit
// before any allocation.
var ErrLimit = errors.New("rmtcmd: limit exceeded")

// keepsSession reports whether a nonzero return code leaves the stream and
// server job in a known state (docs/19, "Command and program calls"). Any
// other code poisons the session because its body semantics are unknown.
func keepsSession(code uint16, command bool) bool {
	switch code {
	case 0x0303, 0x0400:
		return true
	case 0x0401: // invalid explicit command CCSID (level 10 and later)
		return command
	case 0x0500, 0x0501:
		return !command
	}
	return false
}

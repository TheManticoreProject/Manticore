package rmtcmd

import (
	"errors"
	"fmt"
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

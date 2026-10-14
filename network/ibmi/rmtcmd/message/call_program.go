package message

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// RequestCallProgram is the program-call request ID.
const RequestCallProgram uint16 = 0x1003

const callProgramTemplateLength = 23

// CallProgramRequest is request 0x1003.
type CallProgramRequest struct {
	Correlation uint32
	// Program and Library are 10-byte blank-padded names in the job CCSID.
	// The program comes first on the wire.
	Program [10]byte
	Library [10]byte
	// MessageOption is the wire option byte (already level-mapped).
	MessageOption byte
	Parameters    []RequestParameter
}

// Frame encodes the request: a 23-byte template (program, library, message
// option, parameter count), then one envelope per parameter in order.
func (r CallProgramRequest) Frame() (hostserver.Frame, error) {
	if len(r.Parameters) > math.MaxUint16 {
		return hostserver.Frame{}, fmt.Errorf("message: %d parameters exceed %d", len(r.Parameters), math.MaxUint16)
	}
	t := make([]byte, callProgramTemplateLength)
	copy(t[0:], r.Program[:])
	copy(t[10:], r.Library[:])
	t[20] = r.MessageOption
	binary.BigEndian.PutUint16(t[21:], uint16(len(r.Parameters)))
	var v []byte
	for _, p := range r.Parameters {
		var err error
		if v, err = appendParameter(v, p); err != nil {
			return hostserver.Frame{}, err
		}
	}
	return hostserver.Frame{
		Header:   hostserver.Header{ServerID: hostserver.ServerIDRemoteCommand, CorrelationID: r.Correlation, RequestID: RequestCallProgram},
		Template: t,
		Variable: v,
	}, nil
}

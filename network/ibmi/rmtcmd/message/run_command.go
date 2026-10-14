package message

import (
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// RequestRunCommand is the run-command request ID.
const RequestRunCommand uint16 = 0x1002

// Command code points.
const (
	CPCommand          uint16 = 0x1101
	CPCommandWithCCSID uint16 = 0x1104
)

// RunCommandRequest is request 0x1002. Command is raw CL source already
// encoded in the job CCSID.
type RunCommandRequest struct {
	Correlation   uint32
	MessageOption byte
	Command       []byte
	// CommandCCSID, when non-nil (Remote Command level 10 and later),
	// selects CP 0x1104 with this 4-byte CCSID before the command bytes.
	CommandCCSID *uint32
}

// Frame encodes the request: a 1-byte template (message option), then the
// command field.
func (r RunCommandRequest) Frame() (hostserver.Frame, error) {
	var v []byte
	var err error
	if r.CommandCCSID == nil {
		v, err = hostserver.AppendField(nil, CPCommand, r.Command)
	} else {
		value := binary.BigEndian.AppendUint32(make([]byte, 0, 4+len(r.Command)), *r.CommandCCSID)
		v, err = hostserver.AppendField(nil, CPCommandWithCCSID, append(value, r.Command...))
	}
	if err != nil {
		return hostserver.Frame{}, err
	}
	return hostserver.Frame{
		Header:   hostserver.Header{ServerID: hostserver.ServerIDRemoteCommand, CorrelationID: r.Correlation, RequestID: RequestRunCommand},
		Template: []byte{r.MessageOption},
		Variable: v,
	}, nil
}

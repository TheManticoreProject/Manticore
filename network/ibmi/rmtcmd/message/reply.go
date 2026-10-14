package message

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// Reply IDs of the operational requests.
const (
	ReplyRunCommand         uint16 = 0x8002
	ReplyCallProgram        uint16 = 0x8003
	ReplyCallProgramFailure uint16 = 0x9003
)

// OperationReply is the common prefix of 0x8002, 0x8003 and 0x9003: a
// 2-byte return code at offset 20 and a 2-byte message count at offset 22,
// followed at offset 24 by message records, or by output parameters for a
// successful 0x8003 (whose count field is then not meaningful).
type OperationReply struct {
	ReplyID      uint16
	ReturnCode   uint16
	MessageCount uint16
	// Rest is the body from offset 24.
	Rest []byte
}

// ParseOperationReply splits a 0x8002/0x8003/0x9003 reply.
func ParseOperationReply(f hostserver.Frame) (*OperationReply, error) {
	b := body(f)
	if len(b) < 4 {
		return nil, short(fmt.Sprintf("0x%04X reply", f.Header.RequestID), len(b), 4)
	}
	return &OperationReply{
		ReplyID:      f.Header.RequestID,
		ReturnCode:   binary.BigEndian.Uint16(b[0:]),
		MessageCount: binary.BigEndian.Uint16(b[2:]),
		Rest:         b[4:],
	}, nil
}

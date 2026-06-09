package commands

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
)

// CreateRequestCommand creates a request command for the given command code.
//
// Concrete request commands are wired in as they are implemented (one case per
// SMB2 command code), mirroring the SMB 1.0 dispatcher. Until a code's structure
// exists, the command code is reported as unsupported.
func CreateRequestCommand(commandCode codes.CommandCode) (command_interface.CommandInterface, error) {
	switch commandCode {
	// case codes.SMB2_NEGOTIATE:
	// 	return NewNegotiateRequest(), nil
	default:
		return nil, fmt.Errorf("command code not supported: 0x%04x", uint16(commandCode))
	}
}

// CreateResponseCommand creates a response command for the given command code.
//
// Concrete response commands are wired in as they are implemented (one case per
// SMB2 command code), mirroring the SMB 1.0 dispatcher. Until a code's structure
// exists, the command code is reported as unsupported.
func CreateResponseCommand(commandCode codes.CommandCode) (command_interface.CommandInterface, error) {
	switch commandCode {
	// case codes.SMB2_NEGOTIATE:
	// 	return NewNegotiateResponse(), nil
	default:
		return nil, fmt.Errorf("command code not supported: 0x%04x", uint16(commandCode))
	}
}

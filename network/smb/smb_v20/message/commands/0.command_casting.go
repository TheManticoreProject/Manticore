package commands

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
)

// CreateRequestCommand creates a request command for the given command code.
//
// Concrete request commands are wired in as they are implemented (one case per
// SMB2 command code), mirroring the SMB 1.0 dispatcher. Codes without a
// structure yet are reported as unsupported.
func CreateRequestCommand(commandCode codes.CommandCode) (command_interface.CommandInterface, error) {
	switch commandCode {
	case codes.SMB2_NEGOTIATE:
		return NewNegotiateRequest(), nil
	case codes.SMB2_SESSION_SETUP:
		return NewSessionSetupRequest(), nil
	case codes.SMB2_LOGOFF:
		return NewLogoffRequest(), nil
	case codes.SMB2_TREE_CONNECT:
		return NewTreeConnectRequest(), nil
	case codes.SMB2_TREE_DISCONNECT:
		return NewTreeDisconnectRequest(), nil
	case codes.SMB2_ECHO:
		return NewEchoRequest(), nil
	case codes.SMB2_CANCEL:
		return NewCancelRequest(), nil
	case codes.SMB2_FLUSH:
		return NewFlushRequest(), nil
	case codes.SMB2_CLOSE:
		return NewCloseRequest(), nil
	case codes.SMB2_CREATE:
		return NewCreateRequest(), nil
	case codes.SMB2_READ:
		return NewReadRequest(), nil
	case codes.SMB2_WRITE:
		return NewWriteRequest(), nil
	case codes.SMB2_LOCK:
		return NewLockRequest(), nil
	default:
		return nil, fmt.Errorf("command code not supported: 0x%04x", uint16(commandCode))
	}
}

// CreateResponseCommand creates a response command for the given command code.
//
// Concrete response commands are wired in as they are implemented (one case per
// SMB2 command code), mirroring the SMB 1.0 dispatcher. Codes without a
// structure yet are reported as unsupported.
func CreateResponseCommand(commandCode codes.CommandCode) (command_interface.CommandInterface, error) {
	switch commandCode {
	case codes.SMB2_NEGOTIATE:
		return NewNegotiateResponse(), nil
	case codes.SMB2_SESSION_SETUP:
		return NewSessionSetupResponse(), nil
	case codes.SMB2_LOGOFF:
		return NewLogoffResponse(), nil
	case codes.SMB2_TREE_CONNECT:
		return NewTreeConnectResponse(), nil
	case codes.SMB2_TREE_DISCONNECT:
		return NewTreeDisconnectResponse(), nil
	case codes.SMB2_ECHO:
		return NewEchoResponse(), nil
	case codes.SMB2_FLUSH:
		return NewFlushResponse(), nil
	case codes.SMB2_CLOSE:
		return NewCloseResponse(), nil
	case codes.SMB2_CREATE:
		return NewCreateResponse(), nil
	case codes.SMB2_READ:
		return NewReadResponse(), nil
	case codes.SMB2_WRITE:
		return NewWriteResponse(), nil
	case codes.SMB2_LOCK:
		return NewLockResponse(), nil
	default:
		return nil, fmt.Errorf("command code not supported: 0x%04x", uint16(commandCode))
	}
}

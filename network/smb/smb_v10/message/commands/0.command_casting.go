package commands

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
)

// CreateRequestCommand creates a request command for the given command code.
func CreateRequestCommand(commandCode codes.CommandCode) (command_interface.CommandInterface, error) {
	switch commandCode {
	case codes.SMB_COM_NEGOTIATE:
		return NewNegotiateRequest(), nil
	case codes.SMB_COM_LOCKING_ANDX:
		return NewLockingAndxRequest(), nil
	case codes.SMB_COM_LOCK_BYTE_RANGE:
		return NewLockByteRangeRequest(), nil
	case codes.SMB_COM_FLUSH:
		return NewFlushRequest(), nil
	case codes.SMB_COM_CHECK_DIRECTORY:
		return NewCheckDirectoryRequest(), nil
	case codes.SMB_COM_IOCTL:
		return NewIoctlRequest(), nil
	case codes.SMB_COM_CLOSE:
		return NewCloseRequest(), nil
	case codes.SMB_COM_CLOSE_PRINT_FILE:
		return NewClosePrintFileRequest(), nil
	case codes.SMB_COM_CREATE:
		return NewCreateRequest(), nil
	case codes.SMB_COM_CREATE_DIRECTORY:
		return NewCreateDirectoryRequest(), nil
	case codes.SMB_COM_CREATE_NEW:
		return NewCreateNewRequest(), nil
	case codes.SMB_COM_CREATE_TEMPORARY:
		return NewCreateTemporaryRequest(), nil
	case codes.SMB_COM_DELETE:
		return NewDeleteRequest(), nil
	case codes.SMB_COM_DELETE_DIRECTORY:
		return NewDeleteDirectoryRequest(), nil
	case codes.SMB_COM_ECHO:
		return NewEchoRequest(), nil
	case codes.SMB_COM_FIND:
		return NewFindRequest(), nil
	case codes.SMB_COM_FIND_CLOSE:
		return NewFindCloseRequest(), nil
	case codes.SMB_COM_FIND_CLOSE2:
		return NewFindClose2Request(), nil
	case codes.SMB_COM_FIND_UNIQUE:
		return NewFindUniqueRequest(), nil
	case codes.SMB_COM_LOGOFF_ANDX:
		return NewLogoffAndxRequest(), nil
	case codes.SMB_COM_NT_CANCEL:
		return NewNtCancelRequest(), nil
	case codes.SMB_COM_NT_CREATE_ANDX:
		return NewNtCreateAndxRequest(), nil
	case codes.SMB_COM_NT_RENAME:
		return NewNtRenameRequest(), nil
	case codes.SMB_COM_NT_TRANSACT:
		return NewNtTransactRequest(), nil
	case codes.SMB_COM_NT_TRANSACT_SECONDARY:
		return NewNtTransactSecondaryRequest(), nil
	case codes.SMB_COM_OPEN:
		return NewOpenRequest(), nil
	case codes.SMB_COM_OPEN_ANDX:
		return NewOpenAndxRequest(), nil
	case codes.SMB_COM_OPEN_PRINT_FILE:
		return NewOpenPrintFileRequest(), nil
	case codes.SMB_COM_PROCESS_EXIT:
		return NewProcessExitRequest(), nil
	case codes.SMB_COM_QUERY_INFORMATION:
		return NewQueryInformationRequest(), nil
	case codes.SMB_COM_QUERY_INFORMATION2:
		return NewQueryInformation2Request(), nil
	case codes.SMB_COM_QUERY_INFORMATION_DISK:
		return NewQueryInformationDiskRequest(), nil
	case codes.SMB_COM_READ:
		return NewReadRequest(), nil
	case codes.SMB_COM_READ_ANDX:
		return NewReadAndxRequest(), nil
	case codes.SMB_COM_READ_MPX:
		return NewReadMpxRequest(), nil
	case codes.SMB_COM_READ_RAW:
		return NewReadRawRequest(), nil
	case codes.SMB_COM_RENAME:
		return NewRenameRequest(), nil
	case codes.SMB_COM_SEARCH:
		return NewSearchRequest(), nil
	case codes.SMB_COM_SEEK:
		return NewSeekRequest(), nil
	case codes.SMB_COM_SESSION_SETUP_ANDX:
		return NewSessionSetupAndxRequest(), nil
	case codes.SMB_COM_SET_INFORMATION:
		return NewSetInformationRequest(), nil
	case codes.SMB_COM_SET_INFORMATION2:
		return NewSetInformation2Request(), nil
	case codes.SMB_COM_TRANSACTION:
		return NewTransactionRequest(), nil
	case codes.SMB_COM_TRANSACTION_SECONDARY:
		return NewTransactionSecondaryRequest(), nil
	case codes.SMB_COM_TRANSACTION2:
		return NewTransaction2Request(), nil
	case codes.SMB_COM_TRANSACTION2_SECONDARY:
		return NewTransaction2SecondaryRequest(), nil
	case codes.SMB_COM_TREE_CONNECT:
		return NewTreeConnectRequest(), nil
	case codes.SMB_COM_TREE_CONNECT_ANDX:
		return NewTreeConnectAndxRequest(), nil
	case codes.SMB_COM_TREE_DISCONNECT:
		return NewTreeDisconnectRequest(), nil
	case codes.SMB_COM_UNLOCK_BYTE_RANGE:
		return NewUnlockByteRangeRequest(), nil
	case codes.SMB_COM_WRITE:
		return NewWriteRequest(), nil
	case codes.SMB_COM_WRITE_AND_CLOSE:
		return NewWriteAndCloseRequest(), nil
	case codes.SMB_COM_WRITE_AND_UNLOCK:
		return NewWriteAndUnlockRequest(), nil
	case codes.SMB_COM_WRITE_ANDX:
		return NewWriteAndxRequest(), nil
	case codes.SMB_COM_WRITE_MPX:
		return NewWriteMpxRequest(), nil
	case codes.SMB_COM_WRITE_PRINT_FILE:
		return NewWritePrintFileRequest(), nil
	case codes.SMB_COM_WRITE_RAW:
		return NewWriteRawRequest(), nil
	case codes.SMB_COM_LOCK_AND_READ:
		return NewLockAndReadRequest(), nil
	default:
		return nil, fmt.Errorf("command code not supported: %d", commandCode)
	}
}

// CreateResponseCommand creates a response command for the given command code.
func CreateResponseCommand(commandCode codes.CommandCode) (command_interface.CommandInterface, error) {
	switch commandCode {
	case codes.SMB_COM_NEGOTIATE:
		return NewNegotiateResponse(), nil
	case codes.SMB_COM_CHECK_DIRECTORY:
		return NewCheckDirectoryResponse(), nil
	case codes.SMB_COM_CLOSE:
		return NewCloseResponse(), nil
	case codes.SMB_COM_CREATE:
		return NewCreateResponse(), nil
	case codes.SMB_COM_CREATE_DIRECTORY:
		return NewCreateDirectoryResponse(), nil
	case codes.SMB_COM_CREATE_NEW:
		return NewCreateNewResponse(), nil
	case codes.SMB_COM_CREATE_TEMPORARY:
		return NewCreateTemporaryResponse(), nil
	case codes.SMB_COM_DELETE:
		return NewDeleteResponse(), nil
	case codes.SMB_COM_DELETE_DIRECTORY:
		return NewDeleteDirectoryResponse(), nil
	case codes.SMB_COM_ECHO:
		return NewEchoResponse(), nil
	case codes.SMB_COM_FIND:
		return NewFindResponse(), nil
	case codes.SMB_COM_FIND_CLOSE:
		return NewFindCloseResponse(), nil
	case codes.SMB_COM_FIND_CLOSE2:
		return NewFindClose2Response(), nil
	case codes.SMB_COM_FIND_UNIQUE:
		return NewFindUniqueResponse(), nil
	case codes.SMB_COM_FLUSH:
		return NewFlushResponse(), nil
	case codes.SMB_COM_IOCTL:
		return NewIoctlResponse(), nil
	case codes.SMB_COM_LOCK_BYTE_RANGE:
		return NewLockByteRangeResponse(), nil
	case codes.SMB_COM_LOCK_AND_READ:
		return NewLockAndReadResponse(), nil
	case codes.SMB_COM_LOCKING_ANDX:
		return NewLockingAndxResponse(), nil
	case codes.SMB_COM_LOGOFF_ANDX:
		return NewLogoffAndxResponse(), nil
	// case codes.SMB_COM_NT_CANCEL:
	// 	return NewNtCancelResponse(), nil
	case codes.SMB_COM_NT_CREATE_ANDX:
		return NewNtCreateAndxResponse(), nil
	case codes.SMB_COM_NT_RENAME:
		return NewNtRenameResponse(), nil
	case codes.SMB_COM_NT_TRANSACT:
		return NewNtTransactResponse(), nil
	case codes.SMB_COM_NT_TRANSACT_SECONDARY:
		return NewNtTransactSecondaryResponse(), nil
	case codes.SMB_COM_OPEN:
		return NewOpenResponse(), nil
	case codes.SMB_COM_OPEN_ANDX:
		return NewOpenAndxResponse(), nil
	case codes.SMB_COM_OPEN_PRINT_FILE:
		return NewOpenPrintFileResponse(), nil
	case codes.SMB_COM_PROCESS_EXIT:
		return NewProcessExitResponse(), nil
	case codes.SMB_COM_QUERY_INFORMATION:
		return NewQueryInformationResponse(), nil
	case codes.SMB_COM_QUERY_INFORMATION2:
		return NewQueryInformation2Response(), nil
	case codes.SMB_COM_QUERY_INFORMATION_DISK:
		return NewQueryInformationDiskResponse(), nil
	case codes.SMB_COM_READ:
		return NewReadResponse(), nil
	case codes.SMB_COM_READ_ANDX:
		return NewReadAndxResponse(), nil
	case codes.SMB_COM_READ_MPX:
		return NewReadMpxResponse(), nil
	// case codes.SMB_COM_READ_RAW:
	// 	return NewReadRawResponse(), nil
	case codes.SMB_COM_RENAME:
		return NewRenameResponse(), nil
	case codes.SMB_COM_SEARCH:
		return NewSearchResponse(), nil
	case codes.SMB_COM_SEEK:
		return NewSeekResponse(), nil
	case codes.SMB_COM_SESSION_SETUP_ANDX:
		return NewSessionSetupAndxResponse(), nil
	case codes.SMB_COM_SET_INFORMATION:
		return NewSetInformationResponse(), nil
	case codes.SMB_COM_SET_INFORMATION2:
		return NewSetInformation2Response(), nil
	case codes.SMB_COM_TRANSACTION:
		return NewTransactionResponse(), nil
	case codes.SMB_COM_TRANSACTION_SECONDARY:
		return NewTransactionSecondaryResponse(), nil
	case codes.SMB_COM_TRANSACTION2:
		return NewTransaction2Response(), nil
	case codes.SMB_COM_TRANSACTION2_SECONDARY:
		return NewTransaction2SecondaryResponse(), nil
	case codes.SMB_COM_TREE_CONNECT:
		return NewTreeConnectResponse(), nil
	case codes.SMB_COM_TREE_CONNECT_ANDX:
		return NewTreeConnectAndxResponse(), nil
	case codes.SMB_COM_TREE_DISCONNECT:
		return NewTreeDisconnectResponse(), nil
	case codes.SMB_COM_UNLOCK_BYTE_RANGE:
		return NewUnlockByteRangeResponse(), nil
	case codes.SMB_COM_WRITE:
		return NewWriteResponse(), nil
	case codes.SMB_COM_WRITE_AND_CLOSE:
		return NewWriteAndCloseResponse(), nil
	case codes.SMB_COM_WRITE_AND_UNLOCK:
		return NewWriteAndUnlockResponse(), nil
	case codes.SMB_COM_WRITE_ANDX:
		return NewWriteAndxResponse(), nil
	case codes.SMB_COM_WRITE_MPX:
		return NewWriteMpxResponse(), nil
	case codes.SMB_COM_WRITE_PRINT_FILE:
		return NewWritePrintFileResponse(), nil
	case codes.SMB_COM_WRITE_RAW:
		return NewWriteRawFinal(), nil
	case codes.SMB_COM_CLOSE_PRINT_FILE:
		return NewClosePrintFileResponse(), nil
	default:
		return nil, fmt.Errorf("command code not supported: %d", commandCode)
	}
}

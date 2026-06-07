package commands_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
)

// allCommandCodes lists every command code handled by the request/response
// casting registry.
var allCommandCodes = []codes.CommandCode{
	codes.SMB_COM_WRITE_BULK,
	codes.SMB_COM_READ_BULK,
	codes.SMB_COM_CLOSE_AND_TREE_DISC,
	codes.SMB_COM_IOCTL_SECONDARY,
	codes.SMB_COM_QUERY_SERVER,
	codes.SMB_COM_CHECK_DIRECTORY,
	codes.SMB_COM_CLOSE,
	codes.SMB_COM_CLOSE_PRINT_FILE,
	codes.SMB_COM_CREATE,
	codes.SMB_COM_CREATE_DIRECTORY,
	codes.SMB_COM_CREATE_NEW,
	codes.SMB_COM_CREATE_TEMPORARY,
	codes.SMB_COM_DELETE,
	codes.SMB_COM_DELETE_DIRECTORY,
	codes.SMB_COM_ECHO,
	codes.SMB_COM_FIND,
	codes.SMB_COM_FIND_CLOSE,
	codes.SMB_COM_FIND_CLOSE2,
	codes.SMB_COM_FIND_NOTIFY_CLOSE,
	codes.SMB_COM_FIND_UNIQUE,
	codes.SMB_COM_FLUSH,
	codes.SMB_COM_IOCTL,
	codes.SMB_COM_LOCK_AND_READ,
	codes.SMB_COM_LOCK_BYTE_RANGE,
	codes.SMB_COM_LOCKING_ANDX,
	codes.SMB_COM_LOGOFF_ANDX,
	codes.SMB_COM_NEGOTIATE,
	codes.SMB_COM_NEW_FILE_SIZE,
	codes.SMB_COM_NT_CANCEL,
	codes.SMB_COM_NT_CREATE_ANDX,
	codes.SMB_COM_NT_RENAME,
	codes.SMB_COM_NT_TRANSACT,
	codes.SMB_COM_NT_TRANSACT_SECONDARY,
	codes.SMB_COM_OPEN,
	codes.SMB_COM_OPEN_ANDX,
	codes.SMB_COM_OPEN_PRINT_FILE,
	codes.SMB_COM_PROCESS_EXIT,
	codes.SMB_COM_QUERY_INFORMATION,
	codes.SMB_COM_QUERY_INFORMATION2,
	codes.SMB_COM_QUERY_INFORMATION_DISK,
	codes.SMB_COM_READ,
	codes.SMB_COM_READ_ANDX,
	codes.SMB_COM_READ_MPX,
	codes.SMB_COM_READ_RAW,
	codes.SMB_COM_RENAME,
	codes.SMB_COM_SEARCH,
	codes.SMB_COM_SEEK,
	codes.SMB_COM_SESSION_SETUP_ANDX,
	codes.SMB_COM_SET_INFORMATION,
	codes.SMB_COM_SET_INFORMATION2,
	codes.SMB_COM_TRANSACTION,
	codes.SMB_COM_TRANSACTION2,
	codes.SMB_COM_TRANSACTION2_SECONDARY,
	codes.SMB_COM_TRANSACTION_SECONDARY,
	codes.SMB_COM_TREE_CONNECT,
	codes.SMB_COM_TREE_CONNECT_ANDX,
	codes.SMB_COM_TREE_DISCONNECT,
	codes.SMB_COM_UNLOCK_BYTE_RANGE,
	codes.SMB_COM_WRITE,
	codes.SMB_COM_WRITE_AND_CLOSE,
	codes.SMB_COM_WRITE_AND_UNLOCK,
	codes.SMB_COM_WRITE_ANDX,
	codes.SMB_COM_WRITE_MPX,
	codes.SMB_COM_WRITE_PRINT_FILE,
	codes.SMB_COM_WRITE_RAW,
}

// TestUnmarshalOnFreshCommandDoesNotNilPanic verifies that Unmarshal initializes
// its embedded Parameters/Data structures (mirroring Marshal) before using them,
// so calling Unmarshal on a freshly constructed command does not dereference a
// nil pointer. An empty input is used so the call returns early with an error
// instead of exercising unrelated field-parsing paths; without the nil-init
// guard, GetParameters()/GetData() would be nil and the call would panic.
func TestUnmarshalOnFreshCommandDoesNotNilPanic(t *testing.T) {
	producers := map[string]func(codes.CommandCode) (command_interface.CommandInterface, error){
		"request":  commands.CreateRequestCommand,
		"response": commands.CreateResponseCommand,
	}
	for _, code := range allCommandCodes {
		for kind, produce := range producers {
			code, kind, produce := code, kind, produce
			t.Run(kind+"/"+code.String(), func(t *testing.T) {
				cmd, err := produce(code)
				if err != nil || cmd == nil {
					t.Skipf("no %s constructor for %s", kind, code.String())
				}
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("Unmarshal panicked on fresh %s %s: %v", kind, code.String(), r)
					}
				}()
				// Must not panic on a freshly constructed command; an error is fine.
				_, _ = cmd.Unmarshal([]byte{0x00, 0x00, 0x00})
			})
		}
	}
}

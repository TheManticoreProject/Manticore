package codes

import "fmt"

type CommandCode uint8

// SMB Commands
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/5cd5747f-fe0b-40a6-89d0-d67f751f8232
const (
	SMB_COM_CREATE_DIRECTORY       CommandCode = 0x00
	SMB_COM_DELETE_DIRECTORY       CommandCode = 0x01
	SMB_COM_OPEN                   CommandCode = 0x02
	SMB_COM_CREATE                 CommandCode = 0x03
	SMB_COM_CLOSE                  CommandCode = 0x04
	SMB_COM_FLUSH                  CommandCode = 0x05
	SMB_COM_DELETE                 CommandCode = 0x06
	SMB_COM_RENAME                 CommandCode = 0x07
	SMB_COM_QUERY_INFORMATION      CommandCode = 0x08
	SMB_COM_SET_INFORMATION        CommandCode = 0x09
	SMB_COM_READ                   CommandCode = 0x0A
	SMB_COM_WRITE                  CommandCode = 0x0B
	SMB_COM_LOCK_BYTE_RANGE        CommandCode = 0x0C
	SMB_COM_UNLOCK_BYTE_RANGE      CommandCode = 0x0D
	SMB_COM_CREATE_TEMPORARY       CommandCode = 0x0E
	SMB_COM_CREATE_NEW             CommandCode = 0x0F
	SMB_COM_CHECK_DIRECTORY        CommandCode = 0x10
	SMB_COM_PROCESS_EXIT           CommandCode = 0x11
	SMB_COM_SEEK                   CommandCode = 0x12
	SMB_COM_LOCK_AND_READ          CommandCode = 0x13
	SMB_COM_WRITE_AND_UNLOCK       CommandCode = 0x14
	SMB_COM_READ_RAW               CommandCode = 0x1A
	SMB_COM_READ_MPX               CommandCode = 0x1B
	SMB_COM_READ_MPX_SECONDARY     CommandCode = 0x1C
	SMB_COM_WRITE_RAW              CommandCode = 0x1D
	SMB_COM_WRITE_MPX              CommandCode = 0x1E
	SMB_COM_WRITE_MPX_SECONDARY    CommandCode = 0x1F
	SMB_COM_WRITE_COMPLETE         CommandCode = 0x20
	SMB_COM_QUERY_SERVER           CommandCode = 0x21
	SMB_COM_SET_INFORMATION2       CommandCode = 0x22
	SMB_COM_QUERY_INFORMATION2     CommandCode = 0x23
	SMB_COM_LOCKING_ANDX           CommandCode = 0x24
	SMB_COM_TRANSACTION            CommandCode = 0x25
	SMB_COM_TRANSACTION_SECONDARY  CommandCode = 0x26
	SMB_COM_IOCTL                  CommandCode = 0x27
	SMB_COM_IOCTL_SECONDARY        CommandCode = 0x28
	SMB_COM_COPY                   CommandCode = 0x29
	SMB_COM_MOVE                   CommandCode = 0x2A
	SMB_COM_ECHO                   CommandCode = 0x2B
	SMB_COM_WRITE_AND_CLOSE        CommandCode = 0x2C
	SMB_COM_OPEN_ANDX              CommandCode = 0x2D
	SMB_COM_READ_ANDX              CommandCode = 0x2E
	SMB_COM_WRITE_ANDX             CommandCode = 0x2F
	SMB_COM_NEW_FILE_SIZE          CommandCode = 0x30
	SMB_COM_CLOSE_AND_TREE_DISC    CommandCode = 0x31
	SMB_COM_TRANSACTION2           CommandCode = 0x32
	SMB_COM_TRANSACTION2_SECONDARY CommandCode = 0x33
	SMB_COM_FIND_CLOSE2            CommandCode = 0x34
	SMB_COM_FIND_NOTIFY_CLOSE      CommandCode = 0x35
	SMB_COM_TREE_CONNECT           CommandCode = 0x70
	SMB_COM_TREE_DISCONNECT        CommandCode = 0x71
	SMB_COM_NEGOTIATE              CommandCode = 0x72
	SMB_COM_SESSION_SETUP_ANDX     CommandCode = 0x73
	SMB_COM_LOGOFF_ANDX            CommandCode = 0x74
	SMB_COM_TREE_CONNECT_ANDX      CommandCode = 0x75
	SMB_COM_SECURITY_PACKAGE_ANDX  CommandCode = 0x7E
	SMB_COM_QUERY_INFORMATION_DISK CommandCode = 0x80
	SMB_COM_SEARCH                 CommandCode = 0x81
	SMB_COM_FIND                   CommandCode = 0x82
	SMB_COM_FIND_UNIQUE            CommandCode = 0x83
	SMB_COM_FIND_CLOSE             CommandCode = 0x84
	SMB_COM_NT_TRANSACT            CommandCode = 0xA0
	SMB_COM_NT_TRANSACT_SECONDARY  CommandCode = 0xA1
	SMB_COM_NT_CREATE_ANDX         CommandCode = 0xA2
	SMB_COM_NT_CANCEL              CommandCode = 0xA4
	SMB_COM_NT_RENAME              CommandCode = 0xA5
	SMB_COM_OPEN_PRINT_FILE        CommandCode = 0xC0
	SMB_COM_WRITE_PRINT_FILE       CommandCode = 0xC1
	SMB_COM_CLOSE_PRINT_FILE       CommandCode = 0xC2
	SMB_COM_GET_PRINT_QUEUE        CommandCode = 0xC3
	SMB_COM_READ_BULK              CommandCode = 0xD8
	SMB_COM_WRITE_BULK             CommandCode = 0xD9
	SMB_COM_WRITE_BULK_DATA        CommandCode = 0xDA
	SMB_COM_INVALID                CommandCode = 0xFE
	SMB_COM_NO_ANDX_COMMAND        CommandCode = 0xFF
)

var CommandCodeNames = map[CommandCode]string{
	SMB_COM_CREATE_DIRECTORY:       "CREATE_DIRECTORY",
	SMB_COM_DELETE_DIRECTORY:       "DELETE_DIRECTORY",
	SMB_COM_OPEN:                   "OPEN",
	SMB_COM_CREATE:                 "CREATE",
	SMB_COM_CLOSE:                  "CLOSE",
	SMB_COM_FLUSH:                  "FLUSH",
	SMB_COM_DELETE:                 "DELETE",
	SMB_COM_RENAME:                 "RENAME",
	SMB_COM_QUERY_INFORMATION:      "QUERY_INFORMATION",
	SMB_COM_SET_INFORMATION:        "SET_INFORMATION",
	SMB_COM_READ:                   "READ",
	SMB_COM_WRITE:                  "WRITE",
	SMB_COM_LOCK_BYTE_RANGE:        "LOCK_BYTE_RANGE",
	SMB_COM_UNLOCK_BYTE_RANGE:      "UNLOCK_BYTE_RANGE",
	SMB_COM_CREATE_TEMPORARY:       "CREATE_TEMPORARY",
	SMB_COM_CREATE_NEW:             "CREATE_NEW",
	SMB_COM_CHECK_DIRECTORY:        "CHECK_DIRECTORY",
	SMB_COM_PROCESS_EXIT:           "PROCESS_EXIT",
	SMB_COM_SEEK:                   "SEEK",
	SMB_COM_LOCK_AND_READ:          "LOCK_AND_READ",
	SMB_COM_WRITE_AND_UNLOCK:       "WRITE_AND_UNLOCK",
	SMB_COM_READ_RAW:               "READ_RAW",
	SMB_COM_READ_MPX:               "READ_MPX",
	SMB_COM_READ_MPX_SECONDARY:     "READ_MPX_SECONDARY",
	SMB_COM_WRITE_RAW:              "WRITE_RAW",
	SMB_COM_WRITE_MPX:              "WRITE_MPX",
	SMB_COM_WRITE_MPX_SECONDARY:    "WRITE_MPX_SECONDARY",
	SMB_COM_WRITE_COMPLETE:         "WRITE_COMPLETE",
	SMB_COM_QUERY_SERVER:           "QUERY_SERVER",
	SMB_COM_SET_INFORMATION2:       "SET_INFORMATION2",
	SMB_COM_QUERY_INFORMATION2:     "QUERY_INFORMATION2",
	SMB_COM_LOCKING_ANDX:           "LOCKING_ANDX",
	SMB_COM_TRANSACTION:            "TRANSACTION",
	SMB_COM_TRANSACTION_SECONDARY:  "TRANSACTION_SECONDARY",
	SMB_COM_IOCTL:                  "IOCTL",
	SMB_COM_IOCTL_SECONDARY:        "IOCTL_SECONDARY",
	SMB_COM_COPY:                   "COPY",
	SMB_COM_MOVE:                   "MOVE",
	SMB_COM_ECHO:                   "ECHO",
	SMB_COM_WRITE_AND_CLOSE:        "WRITE_AND_CLOSE",
	SMB_COM_OPEN_ANDX:              "OPEN_ANDX",
	SMB_COM_READ_ANDX:              "READ_ANDX",
	SMB_COM_WRITE_ANDX:             "WRITE_ANDX",
	SMB_COM_NEW_FILE_SIZE:          "NEW_FILE_SIZE",
	SMB_COM_CLOSE_AND_TREE_DISC:    "CLOSE_AND_TREE_DISC",
	SMB_COM_TRANSACTION2:           "TRANSACTION2",
	SMB_COM_TRANSACTION2_SECONDARY: "TRANSACTION2_SECONDARY",
	SMB_COM_FIND_CLOSE2:            "FIND_CLOSE2",
	SMB_COM_FIND_NOTIFY_CLOSE:      "FIND_NOTIFY_CLOSE",
	SMB_COM_TREE_CONNECT:           "TREE_CONNECT",
	SMB_COM_TREE_DISCONNECT:        "TREE_DISCONNECT",
	SMB_COM_NEGOTIATE:              "NEGOTIATE",
	SMB_COM_SESSION_SETUP_ANDX:     "SESSION_SETUP_ANDX",
	SMB_COM_LOGOFF_ANDX:            "LOGOFF_ANDX",
	SMB_COM_TREE_CONNECT_ANDX:      "TREE_CONNECT_ANDX",
	SMB_COM_SECURITY_PACKAGE_ANDX:  "SECURITY_PACKAGE_ANDX",
	SMB_COM_QUERY_INFORMATION_DISK: "QUERY_INFORMATION_DISK",
	SMB_COM_SEARCH:                 "SEARCH",
	SMB_COM_FIND:                   "FIND",
	SMB_COM_FIND_UNIQUE:            "FIND_UNIQUE",
	SMB_COM_FIND_CLOSE:             "FIND_CLOSE",
	SMB_COM_NT_TRANSACT:            "NT_TRANSACT",
	SMB_COM_NT_TRANSACT_SECONDARY:  "NT_TRANSACT_SECONDARY",
	SMB_COM_NT_CREATE_ANDX:         "NT_CREATE_ANDX",
	SMB_COM_NT_CANCEL:              "NT_CANCEL",
	SMB_COM_NT_RENAME:              "NT_RENAME",
	SMB_COM_OPEN_PRINT_FILE:        "OPEN_PRINT_FILE",
	SMB_COM_WRITE_PRINT_FILE:       "WRITE_PRINT_FILE",
	SMB_COM_CLOSE_PRINT_FILE:       "CLOSE_PRINT_FILE",
	SMB_COM_GET_PRINT_QUEUE:        "GET_PRINT_QUEUE",
	SMB_COM_READ_BULK:              "READ_BULK",
	SMB_COM_WRITE_BULK:             "WRITE_BULK",
	SMB_COM_WRITE_BULK_DATA:        "WRITE_BULK_DATA",
	SMB_COM_INVALID:                "INVALID",
	SMB_COM_NO_ANDX_COMMAND:        "NO_ANDX_COMMAND",
}

func (c CommandCode) String() string {
	if name, ok := CommandCodeNames[c]; ok {
		return name
	}
	return fmt.Sprintf("CommandCode(%d)", c)
}

package subcommands

// NT Transact Subcommands
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/c2bf4e09-c0e2-42bd-8f7b-6432f1f44d91

type NtTransactSubcommand uint16

const (
	NT_TRANSACT_CREATE              NtTransactSubcommand = 0x0001
	NT_TRANSACT_IOCTL               NtTransactSubcommand = 0x0002
	NT_TRANSACT_SET_SECURITY_DESC   NtTransactSubcommand = 0x0003
	NT_TRANSACT_NOTIFY_CHANGE       NtTransactSubcommand = 0x0004
	NT_TRANSACT_RENAME              NtTransactSubcommand = 0x0005
	NT_TRANSACT_QUERY_SECURITY_DESC NtTransactSubcommand = 0x0006
	NT_TRANSACT_QUERY_QUOTA         NtTransactSubcommand = 0x0007
	NT_TRANSACT_SET_QUOTA           NtTransactSubcommand = 0x0008
)

var NtTransactSubcommandsToString = map[NtTransactSubcommand]string{
	NT_TRANSACT_CREATE:              "CREATE",
	NT_TRANSACT_IOCTL:               "IOCTL",
	NT_TRANSACT_SET_SECURITY_DESC:   "SET_SECURITY_DESC",
	NT_TRANSACT_NOTIFY_CHANGE:       "NOTIFY_CHANGE",
	NT_TRANSACT_RENAME:              "RENAME",
	NT_TRANSACT_QUERY_SECURITY_DESC: "QUERY_SECURITY_DESC",
	NT_TRANSACT_QUERY_QUOTA:         "QUERY_QUOTA",
	NT_TRANSACT_SET_QUOTA:           "SET_QUOTA",
}

func (t NtTransactSubcommand) String() string {
	if str, exists := NtTransactSubcommandsToString[t]; exists {
		return str
	}
	return "UNKNOWN"
}

package subcommands

// Transaction2 Subcommands
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/1cc40e02-aaea-4f33-b7b7-3a6b63906516

type Transaction2Subcommand uint16

const (
	TRANS2_OPEN2                    Transaction2Subcommand = 0x0000
	TRANS2_FIND_FIRST2              Transaction2Subcommand = 0x0001
	TRANS2_FIND_NEXT2               Transaction2Subcommand = 0x0002
	TRANS2_QUERY_FS_INFORMATION     Transaction2Subcommand = 0x0003
	TRANS2_SET_FS_INFORMATION       Transaction2Subcommand = 0x0004
	TRANS2_QUERY_PATH_INFORMATION   Transaction2Subcommand = 0x0005
	TRANS2_SET_PATH_INFORMATION     Transaction2Subcommand = 0x0006
	TRANS2_QUERY_FILE_INFORMATION   Transaction2Subcommand = 0x0007
	TRANS2_SET_FILE_INFORMATION     Transaction2Subcommand = 0x0008
	TRANS2_FSCTL                    Transaction2Subcommand = 0x0009
	TRANS2_IOCTL2                   Transaction2Subcommand = 0x000A
	TRANS2_FIND_NOTIFY_FIRST        Transaction2Subcommand = 0x000B
	TRANS2_FIND_NOTIFY_NEXT         Transaction2Subcommand = 0x000C
	TRANS2_CREATE_DIRECTORY         Transaction2Subcommand = 0x000D
	TRANS2_SESSION_SETUP            Transaction2Subcommand = 0x000E
	TRANS2_GET_DFS_REFERRAL         Transaction2Subcommand = 0x0010
	TRANS2_REPORT_DFS_INCONSISTENCY Transaction2Subcommand = 0x0011
)

var Transaction2SubcommandsToString = map[Transaction2Subcommand]string{
	TRANS2_OPEN2:                    "OPEN2",
	TRANS2_FIND_FIRST2:              "FIND_FIRST2",
	TRANS2_FIND_NEXT2:               "FIND_NEXT2",
	TRANS2_QUERY_FS_INFORMATION:     "QUERY_FS_INFORMATION",
	TRANS2_SET_FS_INFORMATION:       "SET_FS_INFORMATION",
	TRANS2_QUERY_PATH_INFORMATION:   "QUERY_PATH_INFORMATION",
	TRANS2_SET_PATH_INFORMATION:     "SET_PATH_INFORMATION",
	TRANS2_QUERY_FILE_INFORMATION:   "QUERY_FILE_INFORMATION",
	TRANS2_SET_FILE_INFORMATION:     "SET_FILE_INFORMATION",
	TRANS2_FSCTL:                    "FSCTL",
	TRANS2_IOCTL2:                   "IOCTL2",
	TRANS2_FIND_NOTIFY_FIRST:        "FIND_NOTIFY_FIRST",
	TRANS2_FIND_NOTIFY_NEXT:         "FIND_NOTIFY_NEXT",
	TRANS2_CREATE_DIRECTORY:         "CREATE_DIRECTORY",
	TRANS2_SESSION_SETUP:            "SESSION_SETUP",
	TRANS2_GET_DFS_REFERRAL:         "GET_DFS_REFERRAL",
	TRANS2_REPORT_DFS_INCONSISTENCY: "REPORT_DFS_INCONSISTENCY",
}

func (t Transaction2Subcommand) String() string {
	if str, exists := Transaction2SubcommandsToString[t]; exists {
		return str
	}
	return "UNKNOWN"
}

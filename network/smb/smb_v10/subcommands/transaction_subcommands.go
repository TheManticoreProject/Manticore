package subcommands

// Transaction Subcommands
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/227cb147-3c09-4c4b-b145-6c94b04c8231

type TransactionSubcommand uint16

const (
	TRANS_SET_NMPIPE_STATE   TransactionSubcommand = 0x0001
	TRANS_RAW_READ_NMPIPE    TransactionSubcommand = 0x0011
	TRANS_QUERY_NMPIPE_STATE TransactionSubcommand = 0x0021
	TRANS_QUERY_NMPIPE_INFO  TransactionSubcommand = 0x0022
	TRANS_PEEK_NMPIPE        TransactionSubcommand = 0x0023
	TRANS_TRANSACT_NMPIPE    TransactionSubcommand = 0x0026
	TRANS_RAW_WRITE_NMPIPE   TransactionSubcommand = 0x0031
	TRANS_READ_NMPIPE        TransactionSubcommand = 0x0036
	TRANS_WRITE_NMPIPE       TransactionSubcommand = 0x0037
	TRANS_WAIT_NMPIPE        TransactionSubcommand = 0x0053
	TRANS_CALL_NMPIPE        TransactionSubcommand = 0x0054
	TRANS_MAILSLOT_WRITE     TransactionSubcommand = 0x0001
)

var TransactionSubcommandsToString = map[TransactionSubcommand]string{
	TRANS_SET_NMPIPE_STATE:   "SET_NMPIPE_STATE",
	TRANS_RAW_READ_NMPIPE:    "RAW_READ_NMPIPE",
	TRANS_QUERY_NMPIPE_STATE: "QUERY_NMPIPE_STATE",
	TRANS_QUERY_NMPIPE_INFO:  "QUERY_NMPIPE_INFO",
	TRANS_PEEK_NMPIPE:        "PEEK_NMPIPE",
	TRANS_TRANSACT_NMPIPE:    "TRANSACT_NMPIPE",
	TRANS_RAW_WRITE_NMPIPE:   "RAW_WRITE_NMPIPE",
	TRANS_READ_NMPIPE:        "READ_NMPIPE",
	TRANS_WRITE_NMPIPE:       "WRITE_NMPIPE",
	TRANS_WAIT_NMPIPE:        "WAIT_NMPIPE",
	TRANS_CALL_NMPIPE:        "CALL_NMPIPE",
	// TRANS_MAILSLOT_WRITE:     "MAILSLOT_WRITE",
}

func (t TransactionSubcommand) String() string {
	if str, exists := TransactionSubcommandsToString[t]; exists {
		return str
	}
	return "UNKNOWN"
}

package mstsch

// CredFlag enumerates the valid values for the Flags field of TASK_USER_CRED
// ([MS-TSCH] 2.3.9). It is not transmitted on its own; its value populates the 32-bit
// TASK_USER_CRED.Flags field, so it is defined as a uint32 to match that field's width.
type CredFlag uint32

const (
	// CredFlagDefault indicates the credentials are stored using the default mechanism.
	CredFlagDefault CredFlag = 0x00000001
)

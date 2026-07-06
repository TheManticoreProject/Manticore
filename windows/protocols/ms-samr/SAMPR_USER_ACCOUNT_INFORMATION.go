package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_USER_ACCOUNT_INFORMATION holds account-related user attributes
// ([MS-SAMR] 2.2.6.11).
type SAMPR_USER_ACCOUNT_INFORMATION struct {
	UserName           msdtyp.RPC_UNICODE_STRING
	FullName           msdtyp.RPC_UNICODE_STRING
	UserId             ndr.DWORD
	PrimaryGroupId     ndr.DWORD
	HomeDirectory      msdtyp.RPC_UNICODE_STRING
	HomeDirectoryDrive msdtyp.RPC_UNICODE_STRING
	ScriptPath         msdtyp.RPC_UNICODE_STRING
	ProfilePath        msdtyp.RPC_UNICODE_STRING
	AdminComment       msdtyp.RPC_UNICODE_STRING
	WorkStations       msdtyp.RPC_UNICODE_STRING
	LastLogon          OLD_LARGE_INTEGER
	LastLogoff         OLD_LARGE_INTEGER
	LogonHours         SAMPR_LOGON_HOURS
	BadPasswordCount   uint16
	LogonCount         uint16
	PasswordLastSet    OLD_LARGE_INTEGER
	AccountExpires     OLD_LARGE_INTEGER
	UserAccountControl ndr.DWORD
}

package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_USER_LOGON_INFORMATION holds logon-related user attributes ([MS-SAMR]
// 2.2.6.10).
type SAMPR_USER_LOGON_INFORMATION struct {
	UserName           dtyp.RPC_UNICODE_STRING
	FullName           dtyp.RPC_UNICODE_STRING
	UserId             ndr.DWORD
	PrimaryGroupId     ndr.DWORD
	HomeDirectory      dtyp.RPC_UNICODE_STRING
	HomeDirectoryDrive dtyp.RPC_UNICODE_STRING
	ScriptPath         dtyp.RPC_UNICODE_STRING
	ProfilePath        dtyp.RPC_UNICODE_STRING
	WorkStations       dtyp.RPC_UNICODE_STRING
	LastLogon          OLD_LARGE_INTEGER
	LastLogoff         OLD_LARGE_INTEGER
	PasswordLastSet    OLD_LARGE_INTEGER
	PasswordCanChange  OLD_LARGE_INTEGER
	PasswordMustChange OLD_LARGE_INTEGER
	LogonHours         SAMPR_LOGON_HOURS
	BadPasswordCount   uint16
	LogonCount         uint16
	UserAccountControl ndr.DWORD
}

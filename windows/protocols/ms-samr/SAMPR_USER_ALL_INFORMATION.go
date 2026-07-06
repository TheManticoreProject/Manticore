package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_USER_ALL_INFORMATION holds the full set of attributes for a user
// account ([MS-SAMR] 2.2.6.6). The WhichFields bit map indicates which fields
// are valid. LmOwfPassword and NtOwfPassword are carried as RPC_SHORT_BLOB.
type SAMPR_USER_ALL_INFORMATION struct {
	LastLogon            OLD_LARGE_INTEGER
	LastLogoff           OLD_LARGE_INTEGER
	PasswordLastSet      OLD_LARGE_INTEGER
	AccountExpires       OLD_LARGE_INTEGER
	PasswordCanChange    OLD_LARGE_INTEGER
	PasswordMustChange   OLD_LARGE_INTEGER
	UserName             msdtyp.RPC_UNICODE_STRING
	FullName             msdtyp.RPC_UNICODE_STRING
	HomeDirectory        msdtyp.RPC_UNICODE_STRING
	HomeDirectoryDrive   msdtyp.RPC_UNICODE_STRING
	ScriptPath           msdtyp.RPC_UNICODE_STRING
	ProfilePath          msdtyp.RPC_UNICODE_STRING
	AdminComment         msdtyp.RPC_UNICODE_STRING
	WorkStations         msdtyp.RPC_UNICODE_STRING
	UserComment          msdtyp.RPC_UNICODE_STRING
	Parameters           msdtyp.RPC_UNICODE_STRING
	LmOwfPassword        RPC_SHORT_BLOB
	NtOwfPassword        RPC_SHORT_BLOB
	PrivateData          msdtyp.RPC_UNICODE_STRING
	SecurityDescriptor   SAMPR_SR_SECURITY_DESCRIPTOR
	UserId               ndr.DWORD
	PrimaryGroupId       ndr.DWORD
	UserAccountControl   ndr.DWORD
	WhichFields          ndr.DWORD
	LogonHours           SAMPR_LOGON_HOURS
	BadPasswordCount     uint16
	LogonCount           uint16
	CountryCode          uint16
	CodePage             uint16
	LmPasswordPresent    uint8
	NtPasswordPresent    uint8
	PasswordExpired      uint8
	PrivateDataSensitive uint8
}

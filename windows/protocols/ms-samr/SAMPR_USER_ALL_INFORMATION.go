package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
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
	UserName             dtyp.RPC_UNICODE_STRING
	FullName             dtyp.RPC_UNICODE_STRING
	HomeDirectory        dtyp.RPC_UNICODE_STRING
	HomeDirectoryDrive   dtyp.RPC_UNICODE_STRING
	ScriptPath           dtyp.RPC_UNICODE_STRING
	ProfilePath          dtyp.RPC_UNICODE_STRING
	AdminComment         dtyp.RPC_UNICODE_STRING
	WorkStations         dtyp.RPC_UNICODE_STRING
	UserComment          dtyp.RPC_UNICODE_STRING
	Parameters           dtyp.RPC_UNICODE_STRING
	LmOwfPassword        RPC_SHORT_BLOB
	NtOwfPassword        RPC_SHORT_BLOB
	PrivateData          dtyp.RPC_UNICODE_STRING
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

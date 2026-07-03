package msnrpc

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// NETLOGON_CONTROL_DATA_INFORMATION ([MS-NRPC] 2.2.1.7.1) is the switch_type(DWORD) union
// carrying the input data for NetrLogonControl2/2Ex, selected by the FunctionCode ([C706]
// 14.3.8).
//
// The IDL maps FunctionCodes 5, 6, 9, 10 to a trusted-domain name and 8 to a user name;
// because the declarative codec matches one numeric value per `case=` tag, each FunctionCode
// gets its own field of the same [unique][string] type. The [string] wchar_t* arms are NDR
// strings (*ndr.WSTR), not a single UTF-16 code unit.
type NETLOGON_CONTROL_DATA_INFORMATION struct {
	Tag ndr.DWORD `ndr:"switch"`

	// TrustedDomainName arm — FunctionCodes 5, 6, 9, 10 ([string] wchar_t*).
	TrustedDomainName5  *ndr.WSTR `ndr:"case=5,unique"`
	TrustedDomainName6  *ndr.WSTR `ndr:"case=6,unique"`
	TrustedDomainName9  *ndr.WSTR `ndr:"case=9,unique"`
	TrustedDomainName10 *ndr.WSTR `ndr:"case=10,unique"`

	// DebugFlag arm — FunctionCode 65534 (NETLOGON_CONTROL_UNLOAD_NETLOGON_DLL).
	DebugFlag ndr.DWORD `ndr:"case=65534"`

	// UserName arm — FunctionCode 8 ([string] wchar_t*).
	UserName *ndr.WSTR `ndr:"case=8,unique"`
}

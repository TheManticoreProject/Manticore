package msnrpc

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// NL_OUT_CHAIN_SET_CLIENT_ATTRIBUTES_V1 ([MS-NRPC] 2.2.1.3.20) is the [out] payload of
// NetrChainSetClientAttributes (opnum 49).
//
// NOTE: the IDL declares OldDnsHostName as `[unique,string] wchar_t**` — a unique pointer to
// a unique pointer to a string. The declarative NDR codec cannot express a double pointer
// (`**ndr.WSTR` is rejected as an unsupported scalar kind), so it is modelled here as a
// single `*ndr.WSTR`, which omits the outer referent id on the wire. This method is not
// live-validated; correct double-pointer support is deferred to an ndr codec enhancement.
type NL_OUT_CHAIN_SET_CLIENT_ATTRIBUTES_V1 struct {
	HubName           *ndr.WSTR  `ndr:"unique"`
	OldDnsHostName    *ndr.WSTR  `ndr:"unique"`
	SupportedEncTypes *ndr.DWORD `ndr:"unique"`
}

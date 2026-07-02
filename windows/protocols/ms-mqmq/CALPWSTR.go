package msmqmq

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// CALPWSTR is a counted array of wide strings ([MS-MQMQ] 2.2.20): cElems followed by a
// [size_is(cElems)] pointer to that many [string] wchar_t* elements. It is the
// VT_VECTOR|VT_LPWSTR PROPVARIANT arm. Each element is a unique pointer to a wide string.
type CALPWSTR struct {
	CElems ndr.DWORD
	PElems []*ndr.WSTR `ndr:"unique,size_is=CElems"`
}

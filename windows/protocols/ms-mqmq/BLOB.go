package msmqmq

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// BLOB is a counted byte buffer ([MS-MQMQ] 2.2.18): a cbSize count followed by a
// [size_is(cbSize)] pointer to that many octets. It is the VT_BLOB arm of a PROPVARIANT.
type BLOB struct {
	CbSize    ndr.DWORD
	PBlobData []uint8 `ndr:"unique,size_is=CbSize"`
}

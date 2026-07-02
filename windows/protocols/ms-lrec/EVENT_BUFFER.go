package mslrec

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// EVENT_BUFFER holds the block of event data returned by RpcNetEventReceiveData
// ([MS-LREC] 2.2.2.1). Buffer carries one or more NET_EVENT_DATA_HEADER structures
// ([MS-LREC] 2.3.2.2) each followed by its event payload, optionally terminated by a
// NET_EVENT_LOST structure ([MS-LREC] 2.3.2.3) when events were dropped.
//
// The IDL declares Buffer as [size_is(BufferLength)] byte*; under the assumed
// pointer_default(unique) it is a [unique] pointer to a conformant array of bytes, so it
// is modeled as a slice tagged unique + size_is (a referent id, then the array body whose
// maximum_count is BufferLength).
type EVENT_BUFFER struct {
	BufferLength ndr.DWORD
	Buffer       []uint8 `ndr:"unique,size_is=BufferLength"`
}

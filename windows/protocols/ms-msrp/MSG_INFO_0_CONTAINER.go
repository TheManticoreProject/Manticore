package msmsrp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// MSG_INFO_0_CONTAINER is a count-prefixed array of MSG_INFO_0 ([MS-MSRP] 2.2.2.3).
// Buffer is [size_is(EntriesRead)] LPMSG_INFO_0 — a unique pointer to a conformant
// array, so it is modeled as a unique slice sized by EntriesRead.
type MSG_INFO_0_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []MSG_INFO_0 `ndr:"unique,size_is=EntriesRead"`
}

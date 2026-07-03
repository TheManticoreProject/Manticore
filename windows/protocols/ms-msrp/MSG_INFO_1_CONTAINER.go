package msmsrp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// MSG_INFO_1_CONTAINER is a count-prefixed array of MSG_INFO_1 ([MS-MSRP] 2.2.2.4).
// Buffer is [size_is(EntriesRead)] LPMSG_INFO_1 — a unique pointer to a conformant
// array, so it is modeled as a unique slice sized by EntriesRead.
type MSG_INFO_1_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []MSG_INFO_1 `ndr:"unique,size_is=EntriesRead"`
}

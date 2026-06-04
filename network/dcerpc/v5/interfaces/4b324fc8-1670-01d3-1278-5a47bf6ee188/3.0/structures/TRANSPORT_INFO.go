package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TRANSPORT_INFO is the [switch_type(unsigned long)] union of server transport
// information levels ([MS-SRVS] 2.2.3.7). Tag carries the discriminant (the
// level) inline, followed by the selected arm. Each arm is an inline VALUE of
// the matching SERVER_TRANSPORT_INFO_<n> structure.
type TRANSPORT_INFO struct {
	Tag        ndr.DWORD               `ndr:"switch"`
	Transport0 SERVER_TRANSPORT_INFO_0 `ndr:"case=0"`
	Transport1 SERVER_TRANSPORT_INFO_1 `ndr:"case=1"`
	Transport2 SERVER_TRANSPORT_INFO_2 `ndr:"case=2"`
	Transport3 SERVER_TRANSPORT_INFO_3 `ndr:"case=3"`
}

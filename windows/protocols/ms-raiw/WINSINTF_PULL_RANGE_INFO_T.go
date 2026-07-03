package msraiw

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// WINSINTF_PULL_RANGE_INFO_T ([MS-RAIW] 2.2.2.8). PPnr is a reserved LPVOID that MUST
// be ignored; it is modeled as a 4-octet value. This structure is defined by the IDL
// but is not carried by any interface method.
type WINSINTF_PULL_RANGE_INFO_T struct {
	PPnr      ndr.DWORD
	OwnAdd    WINSINTF_ADD_T
	MinVersNo WINSINTF_VERS_NO_T
	MaxVersNo WINSINTF_VERS_NO_T
}

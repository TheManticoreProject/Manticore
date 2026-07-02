package msmqds

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// MQSORTSET is the ordered set of sort keys applied to a directory query result
// ([MS-MQDS] 2.2.14): a count (cCol) and a [size_is(cCol)] pointer to that many MQSORTKEY
// entries.
type MQSORTSET struct {
	CCol ndr.DWORD
	ACol []MQSORTKEY `ndr:"unique,size_is=CCol"`
}

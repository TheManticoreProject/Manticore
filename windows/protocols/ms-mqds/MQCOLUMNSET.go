package msmqds

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// MQCOLUMNSET is the set of property identifiers a query returns for each object
// ([MS-MQDS] 2.2.12): a count (cCol) and a [size_is(cCol)] pointer to that many PROPIDs.
type MQCOLUMNSET struct {
	CCol ndr.DWORD
	ACol []msmqmq.PROPID `ndr:"unique,size_is=CCol"`
}

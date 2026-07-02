package msmqds

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// MQSORTKEY is a single query sort key ([MS-MQDS] 2.2.13): the property to sort on
// (propColumn) and the sort direction (dwOrder, QUERY_SORTASCEND or QUERY_SORTDESCEND).
type MQSORTKEY struct {
	PropColumn ndr.DWORD
	DwOrder    ndr.DWORD
}

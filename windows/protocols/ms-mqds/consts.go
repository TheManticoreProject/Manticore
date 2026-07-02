package msmqds

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// Relational operators for MQPROPERTYRESTRICTION.Rel ([MS-MQDS] 2.2.11): the comparison
// applied between the stored property value and the restriction's value (prval).
const (
	PRLT ndr.DWORD = 0 // less than
	PRLE ndr.DWORD = 1 // less than or equal
	PRGT ndr.DWORD = 2 // greater than
	PRGE ndr.DWORD = 3 // greater than or equal
	PREQ ndr.DWORD = 4 // equal
	PRNE ndr.DWORD = 5 // not equal
)

// Sort orders for MQSORTKEY.DwOrder ([MS-MQDS] 2.2.13).
const (
	QUERY_SORTASCEND  ndr.DWORD = 0 // ascending
	QUERY_SORTDESCEND ndr.DWORD = 1 // descending
)

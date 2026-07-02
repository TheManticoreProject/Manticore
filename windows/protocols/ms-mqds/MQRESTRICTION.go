package msmqds

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// MQRESTRICTION is the set of property restrictions applied to a directory query
// ([MS-MQDS] 2.2.10): a count (cRes) and a [size_is(cRes)] pointer to that many
// MQPROPERTYRESTRICTION entries, combined with logical AND.
type MQRESTRICTION struct {
	CRes      ndr.DWORD
	PaPropRes []MQPROPERTYRESTRICTION `ndr:"unique,size_is=CRes"`
}

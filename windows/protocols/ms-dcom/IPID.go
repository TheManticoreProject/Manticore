package msdcom

import (
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// IPID is an interface pointer identifier ([MS-DCOM] 2.2.16): a GUID that identifies a
// specific interface on an object exporter. It is a typedef of GUID in the IDL, so it is
// carried on the wire as a 16-octet GUID (Data1/Data2/Data3 + Data4[8], 4-octet aligned).
//
// It is modeled on msdtyp.GUID rather than windows/guid.GUID because the latter's trailing
// uint64 does not marshal to the required 16 octets under NDR.
type IPID msdtyp.GUID

// GUID returns the IPID as a windows/guid.GUID for display and comparison.
func (i IPID) GUID() guid.GUID { return msdtyp.GUID(i).GUID() }

// String returns the canonical brace-less GUID string form of the IPID.
func (i IPID) String() string { return msdtyp.GUID(i).String() }

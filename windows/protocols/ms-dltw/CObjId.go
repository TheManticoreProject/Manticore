package msdltw

import (
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// CObjId is an ObjectID ([MS-DLTW] 2.2.2): the identifier assigned to a file when it is
// first tracked, carried on the wire as a single 16-octet GUID (the IDL field _object).
//
// The GUID is modeled on msdtyp.GUID rather than windows/guid.GUID because the latter's
// trailing uint64 does not marshal to the required 16 octets under NDR. Use Object.GUID()
// / msdtyp.NewGUID to convert to and from windows/guid.GUID.
type CObjId struct {
	Object msdtyp.GUID
}

// GUID returns the ObjectID as a windows/guid.GUID for display and comparison.
func (o CObjId) GUID() guid.GUID { return o.Object.GUID() }

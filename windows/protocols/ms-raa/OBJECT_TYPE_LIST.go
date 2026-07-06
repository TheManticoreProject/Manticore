package msraa

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// OBJECT_TYPE_LIST identifies an object-type element in a hierarchy of object types
// ([MS-DTYP] 2.4.x / 2.5.3): the authzr access-check request carries an array of these to
// describe an object and its sub-objects (property sets and properties). It is referenced
// by the authzr IDL but defined in [MS-DTYP], so it is modeled here in the protocol
// package rather than pulled from the IDL.
//
// Wire layout ([C706] 14): Level is a 2-octet WORD, Remaining a 4-octet ACCESS_MASK
// (4-aligned, so 2 octets of padding follow Level), and ObjectType a [unique] pointer to
// a GUID (referent id inline, GUID body deferred). The pointer is unique because
// [MS-DTYP]'s IDL declares pointer_default(unique).
type OBJECT_TYPE_LIST struct {
	Level      uint16
	Remaining  ndr.DWORD
	ObjectType *msdtyp.GUID `ndr:"unique"`
}

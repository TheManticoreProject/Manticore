package msnspi

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// PropertyValue_r ([MS-NSPI] 2.2.5.1). Value is a non-encapsulated PROP_VAL_UNION whose
// discriminant is (ulPropTag & 0x0000FFFF); that discriminant is transmitted inline, so it
// must be set before marshalling (see SetDiscriminant).
type PropertyValue_r struct {
	UlPropTag  ndr.DWORD
	UlReserved ndr.DWORD
	Value      PROP_VAL_UNION
}

// SetDiscriminant derives Value.Tag from UlPropTag (the low 16 bits, the property type), so
// a caller that populated only UlPropTag and the selected arm produces a wire-correct value.
func (p *PropertyValue_r) SetDiscriminant() {
	if p == nil {
		return
	}
	p.Value.Tag = int32(p.UlPropTag & 0x0000FFFF)
}

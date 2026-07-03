package msraa

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// AUTHZR_SECURITY_ATTRIBUTE_V1 is a single claim security attribute ([MS-RAA] 2.2.3.6).
// Value is the IDL's [string][size_is(Length)] WCHAR* attribute name — a [unique] pointer
// to a conformant-varying UTF-16 array (see AUTHZR_SECURITY_ATTRIBUTE_STRING_VALUE);
// Values is the [unique] pointer to the size_is(ValueCount) array of typed values. Length
// and ValueCount are derived from the element counts on marshal.
type AUTHZR_SECURITY_ATTRIBUTE_V1 struct {
	Length     ndr.DWORD
	Value      []uint16 `ndr:"unique,varying,size_is=Length"`
	ValueType  uint16
	Reserved   uint16
	Flags      ndr.DWORD
	ValueCount ndr.DWORD
	Values     []AUTHZR_SECURITY_ATTRIBUTE_V1_VALUE `ndr:"unique,size_is=ValueCount"`
}

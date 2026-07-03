package msraa

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// AUTHZR_SECURITY_ATTRIBUTE_STRING_VALUE is the string arm of a claim attribute value
// ([MS-RAA] 2.2.3.4). Value is the IDL's [string][size_is(Length)] WCHAR* — a [unique]
// pointer to a conformant-varying UTF-16 array whose maximum_count is Length; it carries
// its own offset/actual_count words, so the field is tagged varying. Length is derived
// from the element count on marshal.
type AUTHZR_SECURITY_ATTRIBUTE_STRING_VALUE struct {
	Length ndr.DWORD
	Value  []uint16 `ndr:"unique,varying,size_is=Length"`
}

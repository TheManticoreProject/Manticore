package msnspi

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// PropertyTagArray_r ([MS-NSPI] 2.2.8) is a counted array of property tags.
//
// AulPropTag is an embedded conformant-varying array. The IDL declares it
// [size_is(cValues + 1), length_is(cValues)] DWORD aulPropTag[]: the maximum_count
// (allocation bound) is one greater than the actual_count of transmitted elements, so the
// spec reserves a terminal slot that is never placed on the wire.
//
// KNOWN LIMITATION: the ndr codec cannot express size_is(field+1). The declarative tag
// grammar supports only a sibling field, a "field/N" divisor, or a literal, and for an
// embedded conformant array the hoisted maximum_count is derived from the slice length.
// The ndr.Marshaler escape hatch cannot be used either: this type is referenced through
// [unique] pointers throughout the interface, and a pointer whose pointee implements
// Marshaler short-circuits the codec's referent-id handling (bypassing the referent id and
// faulting on NULL). As a result the marshalled maximum_count equals the actual_count
// (== len(AulPropTag)) rather than actual_count + 1. The transmitted element count is
// correct and the actual_count <= maximum_count invariant holds, so receivers that size
// their allocation from maximum_count and read actual_count elements interoperate; a server
// that strictly re-derives maximum_count as size_is(cValues+1) would observe the off-by-one.
// A faithful fix requires a codec feature (a size_is addend) and is tracked as future work.
type PropertyTagArray_r struct {
	CValues    ndr.DWORD
	AulPropTag []ndr.DWORD `ndr:"conformant,varying,size_is=CValues,length_is=CValues"`
}

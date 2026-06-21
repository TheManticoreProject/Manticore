package structures

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// DRSHandleSize is the wire size of an RPC context handle: a 4-octet attributes field
// plus a 16-octet UUID ([C706] 4.2.16.6, [MS-RPCE] 2.3.2.2).
const DRSHandleSize = 20

// DRS_HANDLE is the drsuapi RPC context handle (the IDL's [context_handle] void*). It is
// transmitted inline as its 20 octets — not as an NDR pointer referent — and aligned to
// 4 octets because it begins with the 4-octet attributes word. It is a Marshaler so the
// codec applies that alignment regardless of what precedes it; the zero value is the
// null handle. IDL_DRSBind returns it and every subsequent drsuapi call passes it back.
type DRS_HANDLE [DRSHandleSize]byte

// Compile-time assertion that DRS_HANDLE encodes itself as NDR.
var _ ndr.Marshaler = (*DRS_HANDLE)(nil)

// IsNull reports whether the handle is the null context handle, i.e. its 16-octet UUID
// (the octets after the 4-octet attributes word) is all zero.
func (h DRS_HANDLE) IsNull() bool {
	for _, b := range h[4:] {
		if b != 0 {
			return false
		}
	}
	return true
}

// AlignmentNDR reports the 4-octet alignment of a context handle.
func (*DRS_HANDLE) AlignmentNDR() int { return 4 }

// MarshalNDR writes the 20 octets of the context handle.
func (h *DRS_HANDLE) MarshalNDR(e *ndr.Encoder) error {
	e.WriteBytes(h[:])
	return nil
}

// UnmarshalNDR reads the 20 octets of the context handle.
func (h *DRS_HANDLE) UnmarshalNDR(d *ndr.Decoder) error {
	b, err := d.ReadBytes(DRSHandleSize)
	if err != nil {
		return fmt.Errorf("drsuapi: context handle: %w", err)
	}
	copy(h[:], b)
	return nil
}

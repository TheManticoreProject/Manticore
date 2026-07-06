package msrpce

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// ContextHandleSize is the wire size of an ept_lookup_handle_t / ndr_context_handle: a
// 4-octet attributes field plus a 16-octet UUID ([C706] section 4.2.16.6).
const ContextHandleSize = 20

// ContextHandle models the ept_lookup_handle_t context handle. A context handle is
// transmitted inline as its 20 octets (not as an NDR pointer referent) and aligned to 4
// octets; it is a Marshaler so the codec applies that alignment ahead of the handle
// regardless of how the preceding tower octet string padded out. The zero value is the
// null handle used on the first ept_map call.
type ContextHandle [ContextHandleSize]byte

// Compile-time assertion that ContextHandle encodes itself as NDR.
var _ ndr.Marshaler = (*ContextHandle)(nil)

// IsNull reports whether the handle is the null context handle, i.e. its 16-octet UUID
// (the octets after the 4-octet attributes word) is all zero. ept_lookup signals the end
// of a paged enumeration by returning a null entry handle.
func (h ContextHandle) IsNull() bool {
	for _, b := range h[4:] {
		if b != 0 {
			return false
		}
	}
	return true
}

// AlignmentNDR reports the 4-octet alignment of a context handle (its first field is the
// 4-octet attributes word).
func (*ContextHandle) AlignmentNDR() int { return 4 }

// MarshalNDR writes the 20 octets of the context handle.
func (h *ContextHandle) MarshalNDR(e *ndr.Encoder) error {
	e.WriteBytes(h[:])
	return nil
}

// UnmarshalNDR reads the 20 octets of the context handle.
func (h *ContextHandle) UnmarshalNDR(d *ndr.Decoder) error {
	b, err := d.ReadBytes(ContextHandleSize)
	if err != nil {
		return fmt.Errorf("epm: context handle: %w", err)
	}
	copy(h[:], b)
	return nil
}

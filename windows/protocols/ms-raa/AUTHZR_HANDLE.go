package msraa

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// AuthzrHandleSize is the wire size of an RPC context handle: a 4-octet attributes field
// plus a 16-octet UUID ([C706] 4.2.16.6, [MS-RPCE] 2.3.2.2).
const AuthzrHandleSize = 20

// AUTHZR_HANDLE is the authzr RPC context handle (the IDL's [context_handle] PVOID,
// [MS-RAA] 2.2.1.1). It is transmitted inline as its 20 octets — not as an NDR pointer
// referent — and aligned to 4 octets because it begins with the 4-octet attributes word.
// It is a Marshaler so the codec applies that alignment regardless of what precedes it;
// the zero value is the null handle. AuthzrInitializeContextFromSid returns it and every
// subsequent authzr call passes it back.
type AUTHZR_HANDLE [AuthzrHandleSize]byte

// Compile-time assertion that AUTHZR_HANDLE encodes itself as NDR.
var _ ndr.Marshaler = (*AUTHZR_HANDLE)(nil)

// IsNull reports whether the handle is the null context handle, i.e. its 16-octet UUID
// (the octets after the 4-octet attributes word) is all zero.
func (h AUTHZR_HANDLE) IsNull() bool {
	for _, b := range h[4:] {
		if b != 0 {
			return false
		}
	}
	return true
}

// AlignmentNDR reports the 4-octet alignment of a context handle.
func (*AUTHZR_HANDLE) AlignmentNDR() int { return 4 }

// MarshalNDR writes the 20 octets of the context handle.
func (h *AUTHZR_HANDLE) MarshalNDR(e *ndr.Encoder) error {
	e.WriteBytes(h[:])
	return nil
}

// UnmarshalNDR reads the 20 octets of the context handle.
func (h *AUTHZR_HANDLE) UnmarshalNDR(d *ndr.Decoder) error {
	b, err := d.ReadBytes(AuthzrHandleSize)
	if err != nil {
		return fmt.Errorf("authzr: context handle: %w", err)
	}
	copy(h[:], b)
	return nil
}

package msrpce

import "github.com/TheManticoreProject/Manticore/windows/guid"

// EptUUID models the uuid_t referent of ept_map's optional [in] object pointer
// (uuid_p_t). Octets holds the 16-octet UUID exactly as it appears on the wire
// (Data1/2/3 little-endian, Data4 big-endian), which the codec emits verbatim as the
// referent body of the full pointer. It is a fixed-octet struct rather than an
// ndr.Marshaler so the codec applies its normal pointer handling (a NULL referent id
// when the field is nil); a Marshaler is intercepted before that and would be
// dereferenced even when nil. guid.GUID is not used directly because its Go layout (a
// trailing uint64) does not map to 16 contiguous octets under reflection.
type EptUUID struct {
	Octets [16]byte
}

// NewEptUUID converts a GUID to its uuid_t octet form.
func NewEptUUID(g guid.GUID) EptUUID {
	var u EptUUID
	copy(u.Octets[:], g.ToBytes())
	return u
}

// GUID converts the octets back to a guid.GUID.
func (u EptUUID) GUID() guid.GUID {
	var g guid.GUID
	g.FromRawBytes(u.Octets[:])
	return g
}

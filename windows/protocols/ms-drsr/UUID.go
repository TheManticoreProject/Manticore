package msdrsr

import "github.com/TheManticoreProject/Manticore/windows/guid"

// UUID is the [MS-DTYP]/[C706] uuid_t as it appears on the wire: 16 octets with
// Data1/2/3 little-endian and Data4 big-endian (guid.GUID.ToBytes order). drsuapi uses
// it for both the IDL GUID and UUID types (UuidDsa, puuidClientDsa, DSNAME.Guid, …).
//
// It is a fixed octet array, NOT an embedded guid.GUID, because guid.GUID's Go layout
// (a trailing uint64) does not reflect to 16 contiguous octets under the NDR codec — it
// would emit 18 bytes in the wrong order. It is a plain struct, NOT an ndr.Marshaler, so
// the codec keeps its normal NULL-referent handling for *UUID pointer fields (a
// Marshaler would be dereferenced even when the pointer is nil). This mirrors the proven
// EptUUID in the endpoint-mapper interface. uuid_t is 4-octet aligned under NDR; every
// UUID field in this interface follows a 4-octet field or starts a struct, so natural
// alignment holds without an explicit override.
type UUID struct {
	Octets [16]byte
}

// UUIDFromGUID converts a guid.GUID to its 16-octet wire form.
func UUIDFromGUID(g guid.GUID) UUID {
	var u UUID
	copy(u.Octets[:], g.ToBytes())
	return u
}

// GUID converts the octets back to a guid.GUID.
func (u UUID) GUID() guid.GUID {
	var g guid.GUID
	g.FromRawBytes(u.Octets[:])
	return g
}

// IsZero reports whether the UUID is the all-zero NULL GUID.
func (u UUID) IsZero() bool { return u == UUID{} }

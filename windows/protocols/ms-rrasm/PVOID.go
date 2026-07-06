package msrrasm

// PVOID is the [MS-DTYP] 2.2.35 PVOID (a pointer-sized value). MS-RRASM uses it only
// for reserved fields carried in opaque buffers (e.g. MIB_IPMCAST_OIF.PvReserved),
// transmitted as a 4-octet value; defined with a fixed size (not a platform-dependent
// uintptr) so the wire layout is identical across GOARCH.
type PVOID = uint32

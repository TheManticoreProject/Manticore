package msrrasm

// HANDLE is the [MS-DTYP] 2.2.16 HANDLE. Within MS-RRASM it appears only inside
// fixed-layout structures carried in opaque buffers (e.g. RASMAN_INFO), where it is
// a 4-octet value in the protocol's 32-bit representation. Defined with a fixed size
// (not a platform-dependent uintptr) so the wire layout is identical across GOARCH.
type HANDLE = uint32

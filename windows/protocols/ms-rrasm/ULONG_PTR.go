package msrrasm

// ULONG_PTR is the [MS-DTYP] 2.2.55 ULONG_PTR. MS-RRASM transmits it as a 4-octet
// value in its 32-bit representation; defined with a fixed size (not a
// platform-dependent uintptr) so the wire layout is identical across GOARCH.
type ULONG_PTR = uint32

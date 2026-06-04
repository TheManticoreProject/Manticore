package structures

// SAMPR_LOGON_HOURS holds a bit map of allowed logon hours ([MS-SAMR]
// 2.2.6.5).
//
// The IDL declares LogonHours as
// [size_is(1260), length_is((UnitsPerWeek+7)/8)] unsigned char*, i.e. a
// [unique] pointer to a conformant-varying byte array. It is modeled here as
// unique,varying because the codec derives both the maximum and actual counts
// from the slice length, which decodes correctly. The literal maximum of 1260
// bytes from the IDL is NOT enforced by this model; callers must keep
// LogonHours within that bound.
type SAMPR_LOGON_HOURS struct {
	UnitsPerWeek uint16
	LogonHours   []byte `ndr:"unique,varying"`
}

package mssamr

// SAMPR_LOGON_HOURS holds a bit map of allowed logon hours ([MS-SAMR]
// 2.2.6.5).
//
// The IDL declares LogonHours as
// [size_is(1260), length_is((UnitsPerWeek+7)/8)] unsigned char*, i.e. a
// [unique] pointer to a conformant-varying byte array whose maximum_count is
// the literal 1260 on the wire, as in NLPR_LOGON_HOURS. The actual_count is
// the slice length; callers keep it at (UnitsPerWeek+7)/8.
type SAMPR_LOGON_HOURS struct {
	UnitsPerWeek uint16
	LogonHours   []byte `ndr:"unique,size_is=1260,varying"`
}

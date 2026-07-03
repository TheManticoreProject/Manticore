package msrprn

// SYSTEMTIME is the [MS-DTYP] 2.3.13 calendar time: eight 16-bit fields. Referenced by
// the job/printer info structures but not defined in the MS-RPRN IDL.
type SYSTEMTIME struct {
	WYear         uint16
	WMonth        uint16
	WDayOfWeek    uint16
	WDay          uint16
	WHour         uint16
	WMinute       uint16
	WSecond       uint16
	WMilliseconds uint16
}

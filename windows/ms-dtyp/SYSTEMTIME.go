package msdtyp

// SYSTEMTIME
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/2fefe8dd-ab48-4e33-a7d5-7171455a9289
type SYSTEMTIME struct {
	// WYear: The year.
	WYear WORD
	// WMonth: The month.
	WMonth WORD
	// WDayOfWeek: The day of the week.
	WDayOfWeek WORD
	// WDay: The day.
	WDay WORD
	// WHour: The hour.
	WHour WORD
	// WMinute: The minute.
	WMinute WORD
	// WSecond: The second.
	WSecond WORD
	// WMilliseconds: The milliseconds.
	WMilliseconds WORD
}

type PSYSTEMTIME *SYSTEMTIME

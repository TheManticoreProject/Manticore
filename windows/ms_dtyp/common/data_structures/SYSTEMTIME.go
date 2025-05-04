package data_structures

import "github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_types"

// SYSTEMTIME
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/2fefe8dd-ab48-4e33-a7d5-7171455a9289
type SYSTEMTIME struct {
	// WYear: The year.
	WYear data_types.WORD
	// WMonth: The month.
	WMonth data_types.WORD
	// WDayOfWeek: The day of the week.
	WDayOfWeek data_types.WORD
	// WDay: The day.
	WDay data_types.WORD
	// WHour: The hour.
	WHour data_types.WORD
	// WMinute: The minute.
	WMinute data_types.WORD
	// WSecond: The second.
	WSecond data_types.WORD
	// WMilliseconds: The milliseconds.
	WMilliseconds data_types.WORD
}

type PSYSTEMTIME *SYSTEMTIME

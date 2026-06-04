package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TIME_OF_DAY_INFO contains the time-of-day information on the server
// ([MS-SRVS] 2.2.4.92). All fields are DWORD except TodTimezone, which is a
// signed long (minutes west of GMT).
type TIME_OF_DAY_INFO struct {
	TodElapsedt  ndr.DWORD
	TodMsecs     ndr.DWORD
	TodHours     ndr.DWORD
	TodMins      ndr.DWORD
	TodSecs      ndr.DWORD
	TodHunds     ndr.DWORD
	TodTimezone  int32
	TodTinterval ndr.DWORD
	TodDay       ndr.DWORD
	TodMonth     ndr.DWORD
	TodYear      ndr.DWORD
	TodWeekday   ndr.DWORD
}

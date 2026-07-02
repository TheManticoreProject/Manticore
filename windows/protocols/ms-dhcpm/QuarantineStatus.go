package msdhcpm

// QuarantineStatus is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type QuarantineStatus uint16

const (
	NOQUARANTINE       QuarantineStatus = 0
	RESTRICTEDACCESS   QuarantineStatus = 1
	DROPPACKET         QuarantineStatus = 2
	PROBATION          QuarantineStatus = 3
	EXEMPT             QuarantineStatus = 4
	DEFAULTQUARSETTING QuarantineStatus = 5
	NOQUARINFO         QuarantineStatus = 6
)

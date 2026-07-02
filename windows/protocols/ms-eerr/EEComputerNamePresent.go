package mseerr

// EEComputerNamePresent is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-EERR]).
type EEComputerNamePresent uint16

const (
	EecnpPresent    EEComputerNamePresent = 1
	EecnpNotPresent EEComputerNamePresent = 2
)

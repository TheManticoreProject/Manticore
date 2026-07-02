package mseerr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// ExtendedErrorInfo models the ExtendedErrorInfo structure ([MS-EERR] 2.2.4): one
// node in a singly linked chain of error records. Next is a [unique] self-pointer
// (nil terminates the chain). Params is a conformant array embedded as the final
// struct member (not a pointer); its maximum_count is hoisted to the head of the
// struct on the wire ([C706] 14.3.7).
type ExtendedErrorInfo struct {
	Next                *ExtendedErrorInfo `ndr:"unique"`
	ComputerName        EEComputerName
	ProcessID           ndr.DWORD
	TimeStamp           int64
	GeneratingComponent ndr.DWORD
	Status              ndr.DWORD
	DetectionLocation   uint16
	Flags               uint16
	NLen                int16
	Params              []ExtendedErrorParam `ndr:"conformant,size_is=NLen"`
}

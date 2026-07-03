package msmqrr

// SectionType is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-MQRR]).
type SectionType uint16

// SectionType discriminant values ([MS-MQRR] section 2.2.x). Exported (unlike the lowercase
// IDL enumerator names stFullPacket, …) so callers of the exported SectionBuffer type can
// branch on SectionBufferType.
const (
	StFullPacket          SectionType = 0
	StBinaryFirstSection  SectionType = 1
	StBinarySecondSection SectionType = 2
	StSrmpFirstSection    SectionType = 3
	StSrmpSecondSection   SectionType = 4
)

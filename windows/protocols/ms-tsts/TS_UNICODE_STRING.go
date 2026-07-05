package mststs

// TS_UNICODE_STRING is the counted wide string embedded in TS_SYS_PROCESS_INFORMATION
// ([MS-TSTS] 2.2.2.7.1, allproc.h _TS_UNICODE_STRING). Unlike RPC_UNICODE_STRING its
// size_is/length_is bounds count wide characters directly (the value of MaximumLength and
// Length), not bytes.
type TS_UNICODE_STRING struct {
	Length        uint16
	MaximumLength uint16
	Buffer        []uint16 `ndr:"unique,size_is=MaximumLength,length_is=Length"`
}

package mststs

// NT6_TS_UNICODE_STRING is the counted wide string embedded in
// TS_SYS_PROCESS_INFORMATION_NT6 ([MS-TSTS] 2.2.2.7.4, allproc.h _NT6_TS_UNICODE_STRING).
//
// Its IDL bounds are size_is(MaximumLength/2)/length_is(Length/2): Length and MaximumLength
// are byte counts and the buffer holds half as many wide characters. The codec expresses
// the /2 divisor directly (see ndr.splitDivisor), so Buffer is a [unique] pointer to a
// conformant-varying array of MaximumLength/2 wide chars, Length/2 of them valid.
type NT6_TS_UNICODE_STRING struct {
	Length        uint16
	MaximumLength uint16
	Buffer        []uint16 `ndr:"unique,size_is=MaximumLength/2,length_is=Length/2"`
}

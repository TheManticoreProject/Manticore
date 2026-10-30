package mseven

// RPC_STRING is a counted ASCII string ([MS-EVEN] 2.2.7). Length and MaximumLength
// are byte counts; Buffer is a [unique] pointer to a conformant-varying char array
// ([size_is(MaximumLength), length_is(Length)] char*): max_count is MaximumLength and
// actual_count is Length, not the slice length.
type RPC_STRING struct {
	Length        uint16
	MaximumLength uint16
	Buffer        []byte `ndr:"unique,varying,size_is=MaximumLength,length_is=Length"`
}

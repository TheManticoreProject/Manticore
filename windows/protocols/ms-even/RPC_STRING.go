package mseven

// RPC_STRING is a counted ASCII string ([MS-EVEN] 2.2.7). Length and MaximumLength
// are byte counts; Buffer is a [unique] pointer to a conformant-varying char array
// ([size_is(MaximumLength), length_is(Length)] char*), so it is tagged unique,varying
// — the NDR max_count/offset/actual_count words travel with the array itself.
type RPC_STRING struct {
	Length        uint16
	MaximumLength uint16
	Buffer        []byte `ndr:"unique,varying"`
}

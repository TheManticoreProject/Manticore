package mssamr

// RPC_STRING is a counted ASCII string ([MS-SAMR] 2.2.3.10). Length and
// MaximumLength are byte counts; Buffer is a [unique] pointer to a
// conformant-varying array of chars.
type RPC_STRING struct {
	Length        uint16
	MaximumLength uint16
	Buffer        []byte `ndr:"unique,varying"`
}

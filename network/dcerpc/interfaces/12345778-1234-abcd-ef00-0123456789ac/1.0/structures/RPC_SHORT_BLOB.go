package structures

// RPC_SHORT_BLOB is a counted array of unsigned short values ([MS-SAMR] 2.2.3.11).
// Length and MaximumLength are byte counts; Buffer is a [unique] pointer to a
// conformant-varying array of unsigned shorts.
type RPC_SHORT_BLOB struct {
	Length        uint16
	MaximumLength uint16
	Buffer        []uint16 `ndr:"unique,varying"`
}

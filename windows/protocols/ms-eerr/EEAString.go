package mseerr

// EEAString models the EEAString structure ([MS-EERR] 2.2.1.1): a length-prefixed ANSI
// (byte) string. PString is a [unique] pointer to a conformant byte array bounded by
// NLength; nil when NLength is 0.
type EEAString struct {
	NLength int16
	PString []uint8 `ndr:"unique,size_is=NLength"`
}

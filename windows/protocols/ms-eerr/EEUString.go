package mseerr

// EEUString models the EEUString structure ([MS-EERR] 2.2.1.2): a length-prefixed
// Unicode (UTF-16) string. PString is a [unique] pointer to a conformant array of
// unsigned shorts bounded by NLength (a character count); nil when NLength is 0.
type EEUString struct {
	NLength int16
	PString []uint16 `ndr:"unique,size_is=NLength"`
}

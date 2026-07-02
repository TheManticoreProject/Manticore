package structures

// BIDI_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-RPRN]).
type BIDI_TYPE uint16

const (
	BIDI_NULL   BIDI_TYPE = 0
	BIDI_INT    BIDI_TYPE = 1
	BIDI_FLOAT  BIDI_TYPE = 2
	BIDI_BOOL   BIDI_TYPE = 3
	BIDI_STRING BIDI_TYPE = 4
	BIDI_TEXT   BIDI_TYPE = 5
	BIDI_ENUM   BIDI_TYPE = 6
	BIDI_BLOB   BIDI_TYPE = 7
)

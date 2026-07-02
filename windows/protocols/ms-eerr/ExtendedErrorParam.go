package mseerr

// ExtendedErrorParam models the ExtendedErrorParam structure ([MS-EERR] 2.2.2): a
// tagged value carried in an ExtendedErrorInfo record. Type is the discriminant enum;
// Field is the non-encapsulated union selected by it.
type ExtendedErrorParam struct {
	Type  ExtendedErrorParamTypesInternal
	Field ExtendedErrorParam_Field
}

// ExtendedErrorParam_Field is the non-encapsulated union of ExtendedErrorParam,
// switch_is(Type) with switch_type(short) ([MS-EERR] 2.2.2, [C706] 14.3.8). The
// discriminant is a 16-bit short (Tag) transmitted inline as the first part of the
// union — so the selecting value appears on the wire twice: once as the enclosing
// ExtendedErrorParam.Type and once here. eeptiNone (6) selects an empty arm (no field).
type ExtendedErrorParam_Field struct {
	Tag           int16        `ndr:"switch"`
	AnsiString    EEAString    `ndr:"case=1"`
	UnicodeString EEUString    `ndr:"case=2"`
	LVal          int32        `ndr:"case=3"`
	IVal          int16        `ndr:"case=4"`
	PVal          int64        `ndr:"case=5"`
	Blob          BinaryEEInfo `ndr:"case=7"`
}

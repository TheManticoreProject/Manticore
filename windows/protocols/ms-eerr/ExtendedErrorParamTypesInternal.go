package mseerr

// ExtendedErrorParamTypesInternal is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-EERR]).
type ExtendedErrorParamTypesInternal uint16

const (
	EeptiAnsiString    ExtendedErrorParamTypesInternal = 1
	EeptiUnicodeString ExtendedErrorParamTypesInternal = 2
	EeptiLongVal       ExtendedErrorParamTypesInternal = 3
	EeptiShortValue    ExtendedErrorParamTypesInternal = 4
	EeptiPointerValue  ExtendedErrorParamTypesInternal = 5
	EeptiNone          ExtendedErrorParamTypesInternal = 6
	EeptiBinary        ExtendedErrorParamTypesInternal = 7
)

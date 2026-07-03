package msnspi

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// PROP_VAL_UNION ([MS-NSPI] 2.2.3.1) is the value component of a PropertyValue_r. It is a
// non-encapsulated union whose discriminant is the low 16 bits of the containing
// PropertyValue_r's ulPropTag (the property type). [C706] section 14.3.8 transmits the
// discriminant inline ahead of the selected arm, so Tag MUST be set to the property type
// (ulPropTag & 0x0000FFFF) before marshalling.
type PROP_VAL_UNION struct {
	Tag    int32           `ndr:"switch"`
	I      int16           `ndr:"case=2"`         // PtypInteger16
	L      int32           `ndr:"case=3"`         // PtypInteger32
	B      uint16          `ndr:"case=11"`        // PtypBoolean (2-octet)
	LpszA  *ndr.STR        `ndr:"case=30,unique"` // PtypString8
	Bin    Binary_r        `ndr:"case=258"`       // PtypBinary
	LpszW  *ndr.WSTR       `ndr:"case=31,unique"` // PtypString
	Lpguid *FlatUID_r      `ndr:"case=72,unique"` // PtypGuid
	Ft     FILETIME        `ndr:"case=64"`        // PtypTime
	Err    int32           `ndr:"case=10"`        // PtypErrorCode
	MVi    ShortArray_r    `ndr:"case=4098"`      // PtypMultipleInteger16
	MVl    LongArray_r     `ndr:"case=4099"`      // PtypMultipleInteger32
	MVszA  StringArray_r   `ndr:"case=4126"`      // PtypMultipleString8
	MVbin  BinaryArray_r   `ndr:"case=4354"`      // PtypMultipleBinary
	MVguid FlatUIDArray_r  `ndr:"case=4168"`      // PtypMultipleGuid
	MVszW  WStringArray_r  `ndr:"case=4127"`      // PtypMultipleString
	MVft   DateTimeArray_r `ndr:"case=4160"`      // PtypMultipleTime
	// The IDL declares lReserved for two discriminants: [case(0x00000001, 0x0000000D)] long
	// lReserved. The codec matches a single case value per field, so each label gets its own
	// arm (both a plain reserved long); the value is identical whichever discriminant selects.
	LReserved   int32 `ndr:"case=1"`  // PtypNull
	LReserved0D int32 `ndr:"case=13"` // 0x0000000D (PtypObject placeholder), same reserved long
}

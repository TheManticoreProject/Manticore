package mseerr

// EEComputerName models the EEComputerName structure ([MS-EERR] 2.2.3): the name of
// the computer where an error was detected. Type indicates whether a name is present;
// Field is the non-encapsulated union selected by it.
type EEComputerName struct {
	Type  EEComputerNamePresent
	Field EEComputerName_Field
}

// EEComputerName_Field is the non-encapsulated union of EEComputerName,
// switch_is(Type) with switch_type(short) ([MS-EERR] 2.2.3, [C706] 14.3.8). The 16-bit
// discriminant (Tag) is transmitted inline as the first part of the union. eecnpPresent
// (1) selects Name; eecnpNotPresent (2) selects an empty arm (no field).
type EEComputerName_Field struct {
	Tag  int16     `ndr:"switch"`
	Name EEUString `ndr:"case=1"`
}

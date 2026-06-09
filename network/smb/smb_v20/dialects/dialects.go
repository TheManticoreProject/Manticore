package dialects

import "fmt"

// Dialect is a 16-bit SMB2 dialect revision number, as carried in the Dialects
// array of an SMB2 NEGOTIATE Request and the DialectRevision field of an SMB2
// NEGOTIATE Response.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/e14db7ff-763a-4263-8b10-0c3944f52fc5
type Dialect uint16

const (
	// SMB2_DIALECT_2_0_2 is the SMB 2.0.2 dialect revision number.
	SMB2_DIALECT_2_0_2 Dialect = 0x0202
	// SMB2_DIALECT_2_1_0 is the SMB 2.1 dialect revision number.
	SMB2_DIALECT_2_1_0 Dialect = 0x0210
	// SMB2_DIALECT_3_0_0 is the SMB 3.0 dialect revision number.
	SMB2_DIALECT_3_0_0 Dialect = 0x0300
	// SMB2_DIALECT_3_0_2 is the SMB 3.0.2 dialect revision number.
	SMB2_DIALECT_3_0_2 Dialect = 0x0302
	// SMB2_DIALECT_3_1_1 is the SMB 3.1.1 dialect revision number.
	SMB2_DIALECT_3_1_1 Dialect = 0x0311
	// SMB2_DIALECT_WILDCARD is the SMB2 wildcard revision number, returned by a
	// server in response to a multi-protocol negotiate with the "SMB 2.???"
	// dialect string.
	SMB2_DIALECT_WILDCARD Dialect = 0x02FF
)

var dialectNames = map[Dialect]string{
	SMB2_DIALECT_2_0_2:    "SMB 2.0.2",
	SMB2_DIALECT_2_1_0:    "SMB 2.1",
	SMB2_DIALECT_3_0_0:    "SMB 3.0",
	SMB2_DIALECT_3_0_2:    "SMB 3.0.2",
	SMB2_DIALECT_3_1_1:    "SMB 3.1.1",
	SMB2_DIALECT_WILDCARD: "SMB 2.???",
}

// String returns the human-readable name of the dialect.
func (d Dialect) String() string {
	if name, ok := dialectNames[d]; ok {
		return name
	}
	return fmt.Sprintf("Dialect(0x%04x)", uint16(d))
}

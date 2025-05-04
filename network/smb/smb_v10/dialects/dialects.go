package dialects

import (
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// SMB Dialects
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/69a29f73-de0c-45a6-a1aa-8ceeea42217f
const (
	// PC NETWORK PROGRAM 1.0 - The original MSNET SMB protocol (otherwise known as the "core protocol")
	DIALECT_PC_NETWORK_PROGRAM_1_0 = "PC NETWORK PROGRAM 1.0"
	// PCLAN1.0 - Some versions of the original MSNET defined this as an alternate to the core protocol name
	DIALECT_PCLAN1_0 = "PCLAN1.0"
	// MICROSOFT NETWORKS 1.03 - This is used for the MS-NET 1.03 product. It defines Lock&Read,Write&Unlock, and a special version of raw read and raw write.
	DIALECT_MICROSOFT_NETWORKS_1_03 = "MICROSOFT NETWORKS 1.03"
	// MICROSOFT NETWORKS 3.0 - This is the DOS LANMAN 1.0 specific protocol. It is equivalent to the LANMAN 1.0 protocol, except the server is required to map errors from the OS/2 error to an appropriate DOS error.
	DIALECT_MICROSOFT_NETWORKS_3_0 = "MICROSOFT NETWORKS 3.0"

	// LANMAN1.0 - This is the first version of the full LANMAN 1.0 protocol
	DIALECT_LANMAN_1_0 = "LANMAN1.0"
	// LANMAN1.2 - This is the second version of the full LANMAN 1.0 protocol
	DIALECT_LANMAN_1_2 = "LANMAN1.2"
	// LANMAN2.0 - LANMAN2.0
	DIALECT_LANMAN_2_0 = "LANMAN2.0"
	// LANMAN2.1 - OS/2 LANMAN2.1
	DIALECT_LANMAN_2_1 = "LANMAN2.1"
	// LM1.2X002 - This is the first version of the full LANMAN 2.0 protocol
	DIALECT_LM1_2X002 = "LM1.2X002"
	// DOS LM1.2X002 - This is the DOS equivalent of the LM1.2X002 protocol. It is identical to the LM1.2X002 protocol, but the server will perform error mapping to appropriate DOS errors.
	DIALECT_DOS_LM1_2X002 = "DOS LM1.2X002"
	// DOS LANMAN2.1 - DOS LANMAN2.1
	DIALECT_DOS_LANMAN_2_1 = "DOS LANMAN2.1"

	// Windows for Workgroups 3.1a - Windows for Workgroups Version 1.0
	DIALECT_WINDOWS_FOR_WORKGROUPS = "Windows for Workgroups 3.1a"
	// NT LM 0.12 - The SMB protocol designed for NT networking. This has special SMBs which duplicate the NT semantics.
	DIALECT_NT_LM_0_12 = "NT LM 0.12"
)

// SMB_Dialect represents a dialect in the SMB protocol
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/69a29f73-de0c-45a6-a1aa-8ceeea42217f
type Dialects struct {
	Dialects []string
}

// NewDialects creates a new Dialects with the specified dialects
//
// This function creates a new Dialects structure with an empty Dialects field.
func NewDialects() *Dialects {
	return &Dialects{
		Dialects: []string{},
	}
}

// AddDialect adds a dialect to the Dialects structure
//
// This function appends the given dialect to the existing dialects in the Dialects structure.
func (d *Dialects) AddDialect(dialect string) {
	// Add a null byte to the end of the dialect string
	d.Dialects = append(d.Dialects, dialect)
}

// Marshal serializes the Dialects structure into a byte slice
//
// This function serializes the Dialects structure into a byte slice. It creates a new byte slice, appends the
// buffer format and the dialect string to it, and returns the resulting byte slice.
func (d *Dialects) Marshal() ([]byte, error) {
	s := types.SMB_STRING{}
	s.SetBufferFormat(types.UCHAR(0x02))
	s.SetString(strings.Join(d.Dialects, "\x00"))

	marshalled_dialects, err := s.Marshal()
	if err != nil {
		return nil, err
	}
	return marshalled_dialects, nil
}

// Unmarshal deserializes a byte slice into the Dialects structure
//
// This function deserializes a byte slice into the Dialects structure. It checks if the input byte slice is
// at least 2 bytes long (to ensure there's enough data for a buffer format and a null terminator).
// It then resets the Dialects field to ensure we're starting fresh.
// It iterates through the input byte slice, reading buffer format, dialect string, and null terminator.
// It appends the dialect string to the Dialects field and returns the number of bytes read and an error if any.
func (d *Dialects) Unmarshal(data []byte) (int, error) {
	if len(data) < 2 { // At minimum, we need 1 byte for BufferFormat and 1 byte for a null terminator
		return 0, fmt.Errorf("data too short to unmarshal SMB_Dialect")
	}

	// Reset dialects to ensure we're starting fresh
	s := types.SMB_STRING{}

	bytesRead, err := s.Unmarshal(data)
	if err != nil {
		return 0, err
	}

	d.Dialects = strings.Split(string(s.Buffer), "\x00")

	return bytesRead, nil
}

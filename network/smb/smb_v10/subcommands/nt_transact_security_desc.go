package subcommands

import (
	"encoding/binary"
	"fmt"
)

// NT_TRANSACT_QUERY_SECURITY_DESC ([MS-CIFS] section 2.2.7.6) and
// NT_TRANSACT_SET_SECURITY_DESC ([MS-CIFS] section 2.2.7.3) let a client read and apply a
// file's self-relative SECURITY_DESCRIPTOR over SMB. Both carry their arguments in the
// NT_Trans_Parameters block (FID + SecurityInformation); the descriptor itself travels as
// the NT_Trans_Data and is treated here as an opaque blob ([MS-DTYP] section 2.4.6).

// SecurityInformation flags ([MS-CIFS] sections 2.2.7.3.1 / 2.2.7.6.1): which parts of the
// security descriptor a query or set operation applies to. They may be OR-ed together.
const (
	OWNER_SECURITY_INFORMATION uint32 = 0x00000001
	GROUP_SECURITY_INFORMATION uint32 = 0x00000002
	DACL_SECURITY_INFORMATION  uint32 = 0x00000004
	SACL_SECURITY_INFORMATION  uint32 = 0x00000008
)

const (
	ntTransactSecurityDescParametersSize    = 8 // FID(2) + Reserved(2) + SecurityInformation(4)
	ntTransactQuerySecDescResponseParamSize = 4 // LengthNeeded(4)
)

// NtTransactSecurityDescParameters is the NT_Trans_Parameters block of an
// NT_TRANSACT_QUERY_SECURITY_DESC request ([MS-CIFS] section 2.2.7.6.1, where the last
// field is named SecurityInfoFields) and of an NT_TRANSACT_SET_SECURITY_DESC request
// ([MS-CIFS] section 2.2.7.3.1, SecurityInformation). The two requests share this layout;
// a set request additionally carries the SECURITY_DESCRIPTOR as its NT_Trans_Data, and a
// query response carries LengthNeeded (see NtTransactQuerySecurityDescResponseParameters)
// plus the descriptor as its NT_Trans_Data.
type NtTransactSecurityDescParameters struct {
	// FID (2 bytes): the open file whose security descriptor is queried or set.
	FID uint16
	// Reserved (2 bytes): MUST be 0x0000.
	Reserved uint16
	// SecurityInformation (4 bytes): the security-descriptor fields to query/set
	// (OWNER/GROUP/DACL/SACL_SECURITY_INFORMATION flags).
	SecurityInformation uint32
}

// Marshal serializes the 8-octet NT_Trans_Parameters block.
func (p *NtTransactSecurityDescParameters) Marshal() ([]byte, error) {
	b := make([]byte, ntTransactSecurityDescParametersSize)
	binary.LittleEndian.PutUint16(b[0:2], p.FID)
	binary.LittleEndian.PutUint16(b[2:4], p.Reserved)
	binary.LittleEndian.PutUint32(b[4:8], p.SecurityInformation)
	return b, nil
}

// Unmarshal parses the 8-octet NT_Trans_Parameters block.
func (p *NtTransactSecurityDescParameters) Unmarshal(data []byte) (int, error) {
	if len(data) < ntTransactSecurityDescParametersSize {
		return 0, fmt.Errorf("subcommands: NT_TRANSACT security-descriptor parameters require %d bytes, got %d", ntTransactSecurityDescParametersSize, len(data))
	}
	p.FID = binary.LittleEndian.Uint16(data[0:2])
	p.Reserved = binary.LittleEndian.Uint16(data[2:4])
	p.SecurityInformation = binary.LittleEndian.Uint32(data[4:8])
	return ntTransactSecurityDescParametersSize, nil
}

// NtTransactQuerySecurityDescResponseParameters is the NT_Trans_Parameters block of an
// NT_TRANSACT_QUERY_SECURITY_DESC response ([MS-CIFS] section 2.2.7.6.2). The descriptor
// itself is returned as the response NT_Trans_Data; if the client's buffer was too small,
// LengthNeeded reports the required size and the NT_Trans_Data is empty.
type NtTransactQuerySecurityDescResponseParameters struct {
	// LengthNeeded (4 bytes): the length, in octets, of the returned (or required) descriptor.
	LengthNeeded uint32
}

// Marshal serializes the 4-octet response parameter block.
func (p *NtTransactQuerySecurityDescResponseParameters) Marshal() ([]byte, error) {
	b := make([]byte, ntTransactQuerySecDescResponseParamSize)
	binary.LittleEndian.PutUint32(b, p.LengthNeeded)
	return b, nil
}

// Unmarshal parses the 4-octet response parameter block.
func (p *NtTransactQuerySecurityDescResponseParameters) Unmarshal(data []byte) (int, error) {
	if len(data) < ntTransactQuerySecDescResponseParamSize {
		return 0, fmt.Errorf("subcommands: NT_TRANSACT_QUERY_SECURITY_DESC response parameters require %d bytes, got %d", ntTransactQuerySecDescResponseParamSize, len(data))
	}
	p.LengthNeeded = binary.LittleEndian.Uint32(data[0:4])
	return ntTransactQuerySecDescResponseParamSize, nil
}

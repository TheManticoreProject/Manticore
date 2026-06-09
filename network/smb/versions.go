package smb

import "fmt"

type SMBProtocolVersion uint16

const (
	SMB_VERSION_1_0 SMBProtocolVersion = 0x0100
	// SMB_VERSION_2_0 (0x0200) is the abstract "SMB 2.0 family" marker. It is NOT
	// a real wire dialect revision; the first concrete SMB2 dialect is SMB 2.0.2.
	// It is retained for backward compatibility — prefer SMB_VERSION_2_0_2.
	SMB_VERSION_2_0 SMBProtocolVersion = 0x0200
	// SMB_VERSION_2_0_2 (0x0202) is the SMB 2.0.2 dialect revision, the lowest
	// real SMB2 DialectRevision carried on the wire.
	SMB_VERSION_2_0_2 SMBProtocolVersion = 0x0202
	SMB_VERSION_2_1   SMBProtocolVersion = 0x0210
	SMB_VERSION_3_0   SMBProtocolVersion = 0x0300
	// SMB_VERSION_3_0_2 (0x0302) is the SMB 3.0.2 dialect revision.
	SMB_VERSION_3_0_2 SMBProtocolVersion = 0x0302
	SMB_VERSION_3_1_1 SMBProtocolVersion = 0x0311
)

func (v SMBProtocolVersion) String() string {
	return fmt.Sprintf("SMB v%d.%d.%d", v>>8&0xF, v>>4&0xF, v&0xF)
}

func (v SMBProtocolVersion) IsSupported() bool {
	switch v {
	case SMB_VERSION_1_0, SMB_VERSION_2_0, SMB_VERSION_2_0_2, SMB_VERSION_2_1,
		SMB_VERSION_3_0, SMB_VERSION_3_0_2, SMB_VERSION_3_1_1:
		return true
	}
	return false
}

func (v SMBProtocolVersion) IsSMB2() bool {
	switch v {
	case SMB_VERSION_2_0, SMB_VERSION_2_0_2, SMB_VERSION_2_1,
		SMB_VERSION_3_0, SMB_VERSION_3_0_2, SMB_VERSION_3_1_1:
		return true
	}
	return false
}

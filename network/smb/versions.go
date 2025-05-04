package smb

import "fmt"

type SMBProtocolVersion uint16

const (
	SMB_VERSION_1_0   SMBProtocolVersion = 0x0100
	SMB_VERSION_2_0   SMBProtocolVersion = 0x0200
	SMB_VERSION_2_1   SMBProtocolVersion = 0x0210
	SMB_VERSION_3_0   SMBProtocolVersion = 0x0300
	SMB_VERSION_3_1_1 SMBProtocolVersion = 0x0311
)

func (v SMBProtocolVersion) String() string {
	return fmt.Sprintf("SMB v%d.%d.%d", v>>8&0xF, v>>4&0xF, v&0xF)
}

func (v SMBProtocolVersion) IsSupported() bool {
	return v == SMB_VERSION_1_0 || v == SMB_VERSION_2_0 || v == SMB_VERSION_2_1 || v == SMB_VERSION_3_0 || v == SMB_VERSION_3_1_1
}

func (v SMBProtocolVersion) IsSMB2() bool {
	return v == SMB_VERSION_2_0 || v == SMB_VERSION_2_1 || v == SMB_VERSION_3_0 || v == SMB_VERSION_3_1_1
}

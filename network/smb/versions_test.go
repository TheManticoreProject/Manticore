package smb

import "testing"

func TestSMBProtocolVersion_String(t *testing.T) {
	tests := []struct {
		name string
		v    SMBProtocolVersion
		want string
	}{
		{name: "SMB 1.0", v: SMB_VERSION_1_0, want: "SMB v1.0.0"},
		{name: "SMB 2.0", v: SMB_VERSION_2_0, want: "SMB v2.0.0"},
		{name: "SMB 2.1", v: SMB_VERSION_2_1, want: "SMB v2.1.0"},
		{name: "SMB 3.0", v: SMB_VERSION_3_0, want: "SMB v3.0.0"},
		{name: "SMB 3.1.1", v: SMB_VERSION_3_1_1, want: "SMB v3.1.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.String(); got != tt.want {
				t.Errorf("SMBProtocolVersion.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

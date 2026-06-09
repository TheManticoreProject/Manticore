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
		{name: "SMB 2.0.2", v: SMB_VERSION_2_0_2, want: "SMB v2.0.2"},
		{name: "SMB 2.1", v: SMB_VERSION_2_1, want: "SMB v2.1.0"},
		{name: "SMB 3.0", v: SMB_VERSION_3_0, want: "SMB v3.0.0"},
		{name: "SMB 3.0.2", v: SMB_VERSION_3_0_2, want: "SMB v3.0.2"},
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

func TestSMBProtocolVersion_Predicates(t *testing.T) {
	all := []SMBProtocolVersion{
		SMB_VERSION_1_0, SMB_VERSION_2_0, SMB_VERSION_2_0_2, SMB_VERSION_2_1,
		SMB_VERSION_3_0, SMB_VERSION_3_0_2, SMB_VERSION_3_1_1,
	}
	for _, v := range all {
		if !v.IsSupported() {
			t.Errorf("%s: IsSupported() = false, want true", v)
		}
	}

	if SMB_VERSION_1_0.IsSMB2() {
		t.Error("SMB 1.0 must not be reported as SMB2")
	}
	for _, v := range []SMBProtocolVersion{SMB_VERSION_2_0_2, SMB_VERSION_2_1, SMB_VERSION_3_0, SMB_VERSION_3_0_2, SMB_VERSION_3_1_1} {
		if !v.IsSMB2() {
			t.Errorf("%s: IsSMB2() = false, want true", v)
		}
	}

	// A value that is not a known dialect must be rejected.
	if (SMBProtocolVersion(0x0201)).IsSupported() {
		t.Error("unknown dialect 0x0201 must not be reported as supported")
	}
}

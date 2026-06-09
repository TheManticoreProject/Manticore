package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb"
)

func TestEngineSupportsSMB2(t *testing.T) {
	supported := []smb.SMBProtocolVersion{smb.SMB_VERSION_2_0, smb.SMB_VERSION_2_0_2, smb.SMB_VERSION_2_1}
	for _, v := range supported {
		if !engineSupportsSMB2(v) {
			t.Errorf("engineSupportsSMB2(%s) = false, want true", v)
		}
	}
	// SMB1 is not an SMB2 dialect; 3.x has no engine yet.
	for _, v := range []smb.SMBProtocolVersion{smb.SMB_VERSION_1_0, smb.SMB_VERSION_3_0, smb.SMB_VERSION_3_0_2, smb.SMB_VERSION_3_1_1} {
		if engineSupportsSMB2(v) {
			t.Errorf("engineSupportsSMB2(%s) = true, want false", v)
		}
	}
}

func TestWantedFamilies(t *testing.T) {
	cases := []struct {
		name            string
		prefs           []smb.SMBProtocolVersion
		wantSMB1, want2 bool
	}{
		{"empty", nil, false, false},
		{"smb1 only", []smb.SMBProtocolVersion{smb.SMB_VERSION_1_0}, true, false},
		{"smb2 only", []smb.SMBProtocolVersion{smb.SMB_VERSION_2_0_2}, false, true},
		{"both", []smb.SMBProtocolVersion{smb.SMB_VERSION_1_0, smb.SMB_VERSION_2_0_2}, true, true},
		{"3.x unsupported -> neither", []smb.SMBProtocolVersion{smb.SMB_VERSION_3_1_1}, false, false},
		{"default preference -> both", defaultPreference(), true, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s1, s2 := wantedFamilies(c.prefs)
			if s1 != c.wantSMB1 || s2 != c.want2 {
				t.Errorf("wantedFamilies(%v) = (%v,%v), want (%v,%v)", c.prefs, s1, s2, c.wantSMB1, c.want2)
			}
		})
	}
}

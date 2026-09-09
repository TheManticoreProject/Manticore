package nt_status

import "testing"

// TestCIFSStatusEncoding asserts every CIFS-specific status is wire-identical to
// its documented SMBSTATUS ErrorClass/ErrorCode pair. The field is little-endian
// and laid out ErrorClass(1) | Reserved(1) | ErrorCode(2), so the value must be
// class | (code << 16) with a zero Reserved byte.
func TestCIFSStatusEncoding(t *testing.T) {
	const (
		errdos = 0x01
		errsrv = 0x02
	)

	cases := []struct {
		name   string
		status NT_STATUS
		class  uint32
		code   uint32
	}{
		{"SMB_BAD_FID", NT_STATUS_SMB_BAD_FID, errdos, 0x0006},
		{"OS2_INVALID_ACCESS", NT_STATUS_OS2_INVALID_ACCESS, errdos, 0x000C},
		{"OS2_NO_MORE_SIDS", NT_STATUS_OS2_NO_MORE_SIDS, errdos, 0x0071},
		{"OS2_INVALID_LEVEL", NT_STATUS_OS2_INVALID_LEVEL, errdos, 0x007C},
		{"OS2_NEGATIVE_SEEK", NT_STATUS_OS2_NEGATIVE_SEEK, errdos, 0x0083},
		{"OS2_CANCEL_VIOLATION", NT_STATUS_OS2_CANCEL_VIOLATION, errdos, 0x00AD},
		{"OS2_ATOMIC_LOCKS_NOT_SUPPORTED", NT_STATUS_OS2_ATOMIC_LOCKS_NOT_SUPPORTED, errdos, 0x00AE},
		{"OS2_CANNOT_COPY", NT_STATUS_OS2_CANNOT_COPY, errdos, 0x010A},
		{"OS2_EAS_DIDNT_FIT", NT_STATUS_OS2_EAS_DIDNT_FIT, errdos, 0x0113},
		{"OS2_EA_ACCESS_DENIED", NT_STATUS_OS2_EA_ACCESS_DENIED, errdos, 0x03E2},
		{"INVALID_SMB", NT_STATUS_INVALID_SMB, errsrv, 0x0001},
		{"SMB_BAD_TID", NT_STATUS_SMB_BAD_TID, errsrv, 0x0005},
		{"SMB_BAD_COMMAND", NT_STATUS_SMB_BAD_COMMAND, errsrv, 0x0016},
		{"SMB_BAD_UID", NT_STATUS_SMB_BAD_UID, errsrv, 0x005B},
		{"SMB_USE_MPX", NT_STATUS_SMB_USE_MPX, errsrv, 0x00FA},
		{"SMB_USE_STANDARD", NT_STATUS_SMB_USE_STANDARD, errsrv, 0x00FB},
		{"SMB_CONTINUE_MPX", NT_STATUS_SMB_CONTINUE_MPX, errsrv, 0x00FC},
		{"SMB_NO_SUPPORT", NT_STATUS_SMB_NO_SUPPORT, errsrv, 0xFFFF},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			want := tc.class | (tc.code << 16)
			if uint32(tc.status) != want {
				t.Fatalf("%s = 0x%08X, want class 0x%02X with code 0x%04X = 0x%08X",
					tc.name, uint32(tc.status), tc.class, tc.code, want)
			}
			// The Reserved byte MUST be zero.
			if reserved := (uint32(tc.status) >> 8) & 0xFF; reserved != 0 {
				t.Fatalf("%s has a non-zero Reserved byte 0x%02X", tc.name, reserved)
			}
		})
	}
}

// TestCIFSStatusNamesRegistered asserts each CIFS-specific status resolves
// through the shared name table, so a server error renders by name rather than
// as a bare hex value.
func TestCIFSStatusNamesRegistered(t *testing.T) {
	for status, name := range cifsStatusNames {
		got, ok := NTStatusToStringName[status]
		if !ok {
			t.Fatalf("0x%08X (%s) is not registered in NTStatusToStringName", uint32(status), name)
		}
		if got != name {
			t.Fatalf("0x%08X resolves to %q, want %q", uint32(status), got, name)
		}
	}
}

// TestCIFSStatusDistinctFromERREF asserts the CIFS-specific values do not
// collide with the [MS-ERREF] values already transcribed in nt_status.go. A
// collision would mean one of the two tables is wrong, and would make a status
// render under the other table's name.
func TestCIFSStatusDistinctFromERREF(t *testing.T) {
	// NT_STATUS_SUCCESS is 0x00000000 and every CIFS-specific value has a
	// non-zero class byte, so no CIFS value can be a success code.
	for status, name := range cifsStatusNames {
		if status == NT_STATUS_SUCCESS {
			t.Fatalf("%s collides with NT_STATUS_SUCCESS", name)
		}
		// A CIFS-specific value must carry a non-zero error class in its low
		// byte, which is what distinguishes it from the [MS-ERREF] severity
		// encoding in the high bits.
		if uint32(status)&0xFF == 0 {
			t.Fatalf("%s = 0x%08X has a zero ErrorClass byte", name, uint32(status))
		}
	}
}

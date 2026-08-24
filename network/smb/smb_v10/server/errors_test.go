package server

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// TestSMBStatusEncode asserts the SMBSTATUS pair is packed the way [MS-CIFS]
// lays the Status field out: ErrorClass in the low byte, a zero Reserved byte,
// and ErrorCode in the high half-word.
func TestSMBStatusEncode(t *testing.T) {
	cases := []struct {
		name string
		in   SMBStatus
		want uint32
	}{
		{"success", SMBStatus{ERRSUCCESS, 0x0000}, 0x00000000},
		{"ERRDOS/ERRbadfile", SMBStatus{ERRDOS, 0x0002}, 0x00020001},
		{"ERRSRV/ERRinvtid", SMBStatus{ERRSRV, 0x0005}, 0x00050002},
		{"ERRSRV/ERRbaduid", SMBStatus{ERRSRV, 0x005B}, 0x005B0002},
		{"ERRSRV/ERRnosupport", SMBStatus{ERRSRV, 0xFFFF}, 0xFFFF0002},
		{"ERRHRD/ERRdiskfull", SMBStatus{ERRHRD, 0x0027}, 0x00270003},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.in.Encode(); got != tc.want {
				t.Fatalf("Encode() = 0x%08X, want 0x%08X", got, tc.want)
			}
			if reserved := (tc.in.Encode() >> 8) & 0xFF; reserved != 0 {
				t.Fatalf("Encode() left a non-zero Reserved byte 0x%02X", reserved)
			}
		})
	}
}

// TestDOSErrorMappings asserts the tabulated NTSTATUS-to-SMBSTATUS mappings a
// server actually emits, against the [MS-CIFS] 2.2.2.4 table.
func TestDOSErrorMappings(t *testing.T) {
	cases := []struct {
		status nt_status.NT_STATUS
		want   SMBStatus
	}{
		{nt_status.NT_STATUS_SUCCESS, SMBStatus{ERRSUCCESS, 0x0000}},
		{nt_status.NT_STATUS_NOT_IMPLEMENTED, SMBStatus{ERRDOS, 0x0001}},
		{nt_status.NT_STATUS_NO_SUCH_FILE, SMBStatus{ERRDOS, 0x0002}},
		{nt_status.NT_STATUS_OBJECT_NAME_NOT_FOUND, SMBStatus{ERRDOS, 0x0002}},
		{nt_status.NT_STATUS_OBJECT_PATH_NOT_FOUND, SMBStatus{ERRDOS, 0x0003}},
		{nt_status.NT_STATUS_ACCESS_DENIED, SMBStatus{ERRDOS, 0x0005}},
		// STATUS_LOGON_FAILURE is ERRDOS/ERRnoaccess, not an ERRSRV code.
		{nt_status.NT_STATUS_LOGON_FAILURE, SMBStatus{ERRDOS, 0x0005}},
		{nt_status.NT_STATUS_INVALID_HANDLE, SMBStatus{ERRDOS, 0x0006}},
		// STATUS_DIRECTORY_NOT_EMPTY is ERRremcd (0x0010), the code whose
		// description names the condition.
		{nt_status.NT_STATUS_DIRECTORY_NOT_EMPTY, SMBStatus{ERRDOS, 0x0010}},
		{nt_status.NT_STATUS_NO_MORE_FILES, SMBStatus{ERRDOS, 0x0012}},
		{nt_status.NT_STATUS_SHARING_VIOLATION, SMBStatus{ERRDOS, 0x0020}},
		{nt_status.NT_STATUS_LOCK_NOT_GRANTED, SMBStatus{ERRDOS, 0x0021}},
		{nt_status.NT_STATUS_END_OF_FILE, SMBStatus{ERRDOS, 0x0026}},
		{nt_status.NT_STATUS_NOT_SUPPORTED, SMBStatus{ERRDOS, 0x0032}},
		{nt_status.NT_STATUS_OBJECT_NAME_COLLISION, SMBStatus{ERRDOS, 0x0050}},
		{nt_status.NT_STATUS_INVALID_PARAMETER, SMBStatus{ERRDOS, 0x0057}},
		{nt_status.NT_STATUS_INVALID_INFO_CLASS, SMBStatus{ERRDOS, 0x00E6}},
		{nt_status.NT_STATUS_MORE_PROCESSING_REQUIRED, SMBStatus{ERRDOS, 0x00EA}},
		{nt_status.NT_STATUS_NETWORK_ACCESS_DENIED, SMBStatus{ERRSRV, 0x0004}},
		{nt_status.NT_STATUS_BAD_NETWORK_NAME, SMBStatus{ERRSRV, 0x0006}},
		{nt_status.NT_STATUS_BAD_DEVICE_TYPE, SMBStatus{ERRSRV, 0x0007}},
		{nt_status.NT_STATUS_TOO_MANY_SESSIONS, SMBStatus{ERRSRV, 0x005A}},
		{nt_status.NT_STATUS_DISK_FULL, SMBStatus{ERRHRD, 0x0027}},
		{nt_status.NT_STATUS_MEDIA_WRITE_PROTECTED, SMBStatus{ERRHRD, 0x0013}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(statusName(tc.status), func(t *testing.T) {
			got, ok := DOSError(tc.status)
			if !ok {
				t.Fatalf("DOSError(%s) reported no mapping", statusName(tc.status))
			}
			if got != tc.want {
				t.Fatalf("DOSError(%s) = %s/0x%04X, want %s/0x%04X",
					statusName(tc.status), got.Class, got.Code, tc.want.Class, tc.want.Code)
			}
		})
	}
}

// TestDOSErrorDecomposesCIFSStatus asserts a CIFS-specific NTSTATUS is
// decomposed into its class/code pair rather than falling through to the
// unmapped answer: those values are defined to be wire-identical to the pair, so
// no lookup entry is needed.
func TestDOSErrorDecomposesCIFSStatus(t *testing.T) {
	cases := []struct {
		status nt_status.NT_STATUS
		want   SMBStatus
	}{
		{nt_status.NT_STATUS_INVALID_SMB, SMBStatus{ERRSRV, 0x0001}},
		{nt_status.NT_STATUS_SMB_BAD_TID, SMBStatus{ERRSRV, 0x0005}},
		{nt_status.NT_STATUS_SMB_BAD_COMMAND, SMBStatus{ERRSRV, 0x0016}},
		{nt_status.NT_STATUS_SMB_BAD_UID, SMBStatus{ERRSRV, 0x005B}},
		{nt_status.NT_STATUS_SMB_BAD_FID, SMBStatus{ERRDOS, 0x0006}},
		{nt_status.NT_STATUS_OS2_INVALID_LEVEL, SMBStatus{ERRDOS, 0x007C}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(statusName(tc.status), func(t *testing.T) {
			got, ok := DOSError(tc.status)
			if !ok {
				t.Fatalf("DOSError(%s) reported no mapping", statusName(tc.status))
			}
			if got != tc.want {
				t.Fatalf("DOSError(%s) = %s/0x%04X, want %s/0x%04X",
					statusName(tc.status), got.Class, got.Code, tc.want.Class, tc.want.Code)
			}
			// Round trip: re-encoding the decomposed pair must reproduce the
			// original NTSTATUS exactly, which is what "wire-identical" means.
			if encoded := got.Encode(); encoded != uint32(tc.status) {
				t.Fatalf("re-encoding gives 0x%08X, want the original 0x%08X", encoded, uint32(tc.status))
			}
		})
	}
}

// TestDOSErrorUnmapped asserts an NTSTATUS with no legacy equivalent reports
// that fact and still yields something valid to send.
func TestDOSErrorUnmapped(t *testing.T) {
	// A plausible NTSTATUS that the CIFS table does not tabulate.
	got, ok := DOSError(nt_status.NT_STATUS_INTERNAL_ERROR)
	if ok {
		t.Fatalf("DOSError(NT_STATUS_INTERNAL_ERROR) unexpectedly reports a tabulated mapping (%s/0x%04X)", got.Class, got.Code)
	}
	if got != unmappedError {
		t.Fatalf("unmapped status = %s/0x%04X, want ERRSRV/0x0041", got.Class, got.Code)
	}
}

// TestEncodeStatus asserts the header Status field is rendered in whichever
// encoding the client selected: the NTSTATUS unchanged when it negotiated NT
// status codes, and the legacy pair otherwise.
func TestEncodeStatus(t *testing.T) {
	cases := []struct {
		name       string
		status     nt_status.NT_STATUS
		ntStatus   bool
		wantStatus uint32
	}{
		{"success with NT status", nt_status.NT_STATUS_SUCCESS, true, 0x00000000},
		{"success without NT status", nt_status.NT_STATUS_SUCCESS, false, 0x00000000},
		{"access denied with NT status", nt_status.NT_STATUS_ACCESS_DENIED, true, 0xC0000022},
		{"access denied without NT status", nt_status.NT_STATUS_ACCESS_DENIED, false, 0x00050001},
		{"not implemented with NT status", nt_status.NT_STATUS_NOT_IMPLEMENTED, true, 0xC0000002},
		{"not implemented without NT status", nt_status.NT_STATUS_NOT_IMPLEMENTED, false, 0x00010001},
		// A CIFS-specific status is the same 32 bits either way, which is the
		// whole point of the encoding.
		{"bad TID with NT status", nt_status.NT_STATUS_SMB_BAD_TID, true, 0x00050002},
		{"bad TID without NT status", nt_status.NT_STATUS_SMB_BAD_TID, false, 0x00050002},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := EncodeStatus(tc.status, tc.ntStatus); got != tc.wantStatus {
				t.Fatalf("EncodeStatus(%s, %t) = 0x%08X, want 0x%08X",
					statusName(tc.status), tc.ntStatus, got, tc.wantStatus)
			}
		})
	}
}

// TestErrorClassString asserts each class renders by its [MS-CIFS] name.
func TestErrorClassString(t *testing.T) {
	cases := map[ErrorClass]string{
		ERRSUCCESS:       "SUCCESS",
		ERRDOS:           "ERRDOS",
		ERRSRV:           "ERRSRV",
		ERRHRD:           "ERRHRD",
		ERRCMD:           "ERRCMD",
		ErrorClass(0x7F): "UNKNOWN",
	}
	for class, want := range cases {
		if got := class.String(); got != want {
			t.Fatalf("ErrorClass(0x%02X).String() = %q, want %q", uint8(class), got, want)
		}
	}
}

// TestStatusName asserts a status renders by name when one is known and as a hex
// value otherwise, so a log line is never just an opaque number where a name
// exists.
func TestStatusName(t *testing.T) {
	if got := statusName(nt_status.NT_STATUS_ACCESS_DENIED); got != "ACCESS_DENIED" {
		t.Fatalf("statusName(ACCESS_DENIED) = %q", got)
	}
	// The CIFS-specific codes are registered too.
	if got := statusName(nt_status.NT_STATUS_SMB_BAD_TID); got != "SMB_BAD_TID" {
		t.Fatalf("statusName(SMB_BAD_TID) = %q", got)
	}
	if got := statusName(nt_status.NT_STATUS(0xDEADBEEF)); got != "0xdeadbeef" {
		t.Fatalf("statusName(0xDEADBEEF) = %q, want the hex form", got)
	}
}

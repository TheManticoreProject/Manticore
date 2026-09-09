package nt_status_test

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// undefined is a value neither [MS-ERREF] nor [MS-CIFS] defines, so it stands
// for a status the table cannot name.
const undefined nt_status.NT_STATUS = 0xDEADBEEF

func TestString(t *testing.T) {
	tests := []struct {
		status nt_status.NT_STATUS
		want   string
	}{
		{nt_status.NT_STATUS_SUCCESS, "NT_STATUS_SUCCESS"},
		{nt_status.NT_STATUS_PENDING, "NT_STATUS_PENDING"},
		{nt_status.NT_STATUS_TIMEOUT, "NT_STATUS_TIMEOUT"},
		{nt_status.NT_STATUS_BUFFER_ALL_ZEROS, "NT_STATUS_BUFFER_ALL_ZEROS"},
		{nt_status.NT_STATUS_ACCESS_DENIED, "NT_STATUS_ACCESS_DENIED"},
		{nt_status.NT_STATUS_SMB_BAD_UID, "NT_STATUS_SMB_BAD_UID"},
		// An unnamed status renders as hex rather than as a placeholder, so a
		// log line always carries the value.
		{undefined, "0xdeadbeef"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("NT_STATUS(0x%08x).String() = %q, want %q", uint32(tt.status), got, tt.want)
		}
	}
}

func TestStatusValuesMatchTheSpecification(t *testing.T) {
	tests := []struct {
		name   string
		status nt_status.NT_STATUS
		want   uint32
	}{
		{"NT_STATUS_SUCCESS", nt_status.NT_STATUS_SUCCESS, 0x00000000},
		{"NT_STATUS_WAIT_2", nt_status.NT_STATUS_WAIT_2, 0x00000002},
		{"NT_STATUS_PENDING", nt_status.NT_STATUS_PENDING, 0x00000103},
		{"NT_STATUS_BUFFER_OVERFLOW", nt_status.NT_STATUS_BUFFER_OVERFLOW, 0x80000005},
		{"NT_STATUS_ACCESS_DENIED", nt_status.NT_STATUS_ACCESS_DENIED, 0xC0000022},
		{"NT_STATUS_MORE_PROCESSING_REQUIRED", nt_status.NT_STATUS_MORE_PROCESSING_REQUIRED, 0xC0000016},
		{"NT_STATUS_LOGON_FAILURE", nt_status.NT_STATUS_LOGON_FAILURE, 0xC000006D},
		// The [MS-CIFS] extensions keep their ErrorClass | ErrorCode<<16 form.
		{"NT_STATUS_SMB_BAD_FID", nt_status.NT_STATUS_SMB_BAD_FID, 0x00060001},
		{"NT_STATUS_SMB_NO_SUPPORT", nt_status.NT_STATUS_SMB_NO_SUPPORT, 0xFFFF0002},
	}

	for _, tt := range tests {
		if uint32(tt.status) != tt.want {
			t.Errorf("%s = 0x%08X, want 0x%08X", tt.name, uint32(tt.status), tt.want)
		}
	}
}

func TestNameAndDescription(t *testing.T) {
	if got, want := nt_status.NT_STATUS_ACCESS_DENIED.Name(), "NT_STATUS_ACCESS_DENIED"; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
	if got := nt_status.NT_STATUS_ACCESS_DENIED.Description(); !strings.Contains(got, "access") && !strings.Contains(got, "Access") {
		t.Errorf("Description() = %q, want it to describe an access failure", got)
	}

	// A [MS-CIFS] extension reports its own name and description, not those of
	// the [MS-ERREF] value it may share a number with.
	if got, want := nt_status.NT_STATUS_SMB_BAD_TID.Name(), "NT_STATUS_SMB_BAD_TID"; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}

	if got := undefined.Name(); got != "" {
		t.Errorf("an undefined status reported the name %q, want the empty string", got)
	}
	if got := undefined.Description(); got != "" {
		t.Errorf("an undefined status reported the description %q, want the empty string", got)
	}
}

// TestIsSuccessIsNarrowerThanSeverity pins that IsSuccess means the single value
// NT_STATUS_SUCCESS, not the whole success severity, so it agrees with Error.
func TestIsSuccessIsNarrowerThanSeverity(t *testing.T) {
	if !nt_status.NT_STATUS_SUCCESS.IsSuccess() {
		t.Error("NT_STATUS_SUCCESS did not report success")
	}
	for _, status := range []nt_status.NT_STATUS{
		nt_status.NT_STATUS_PENDING,         // severity Success
		nt_status.NT_STATUS_BUFFER_OVERFLOW, // severity Warning
		nt_status.NT_STATUS_ACCESS_DENIED,   // severity Error
		nt_status.NT_STATUS_SMB_BAD_UID,     // an [MS-CIFS] extension
		undefined,
	} {
		if status.IsSuccess() {
			t.Errorf("0x%08x reported success", uint32(status))
		}
		if status.Error() == nil {
			t.Errorf("0x%08x reported success through Error()", uint32(status))
		}
	}
}

func TestError(t *testing.T) {
	if err := nt_status.NT_STATUS_SUCCESS.Error(); err != nil {
		t.Errorf("NT_STATUS_SUCCESS.Error() = %v, want nil", err)
	}

	err := nt_status.NT_STATUS_ACCESS_DENIED.Error()
	if err == nil {
		t.Fatal("NT_STATUS_ACCESS_DENIED.Error() returned nil")
	}
	for _, want := range []string{"0xc0000022", "NT_STATUS_ACCESS_DENIED"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Error() = %q, want it to contain %q", err.Error(), want)
		}
	}

	undefinedErr := undefined.Error()
	if undefinedErr == nil {
		t.Fatal("an undefined non-success status reported nil, want an error")
	}
	if !strings.Contains(undefinedErr.Error(), "0xdeadbeef") {
		t.Errorf("Error() = %q, want it to contain the hexadecimal value", undefinedErr.Error())
	}
}

func TestLookup(t *testing.T) {
	entry, defined := nt_status.Lookup(nt_status.NT_STATUS_INVALID_PARAMETER)
	if !defined {
		t.Fatal("NT_STATUS_INVALID_PARAMETER is not in the table")
	}
	if entry.Name != "NT_STATUS_INVALID_PARAMETER" {
		t.Errorf("entry.Name = %q, want %q", entry.Name, "NT_STATUS_INVALID_PARAMETER")
	}
	if entry.Description == "" {
		t.Error("entry.Description is empty")
	}

	if _, defined := nt_status.Lookup(undefined); defined {
		t.Error("an undefined status was reported as defined")
	}
}

func TestFromName(t *testing.T) {
	tests := []struct {
		name string
		want nt_status.NT_STATUS
	}{
		{"NT_STATUS_SUCCESS", nt_status.NT_STATUS_SUCCESS},
		{"NT_STATUS_ACCESS_DENIED", nt_status.NT_STATUS_ACCESS_DENIED},
		// The name the specification itself uses resolves too, so a status
		// copied out of [MS-ERREF] or a capture needs no translation.
		{"STATUS_SUCCESS", nt_status.NT_STATUS_SUCCESS},
		{"STATUS_ACCESS_DENIED", nt_status.NT_STATUS_ACCESS_DENIED},
		{"RPC_NT_BAD_STUB_DATA", nt_status.NT_STATUS_RPC_NT_BAD_STUB_DATA},
		{"NT_STATUS_SMB_BAD_FID", nt_status.NT_STATUS_SMB_BAD_FID},
	}

	for _, tt := range tests {
		got, defined := nt_status.FromName(tt.name)
		if !defined {
			t.Errorf("FromName(%q) reported the name as undefined", tt.name)
			continue
		}
		if got != tt.want {
			t.Errorf("FromName(%q) = 0x%08X, want 0x%08X", tt.name, uint32(got), uint32(tt.want))
		}
	}

	for _, name := range []string{"", "NT_STATUS_NOT_A_REAL_STATUS", "nt_status_access_denied", "ERROR_ACCESS_DENIED"} {
		if status, defined := nt_status.FromName(name); defined {
			t.Errorf("FromName(%q) resolved to 0x%08X, want undefined", name, uint32(status))
		}
	}
}

// TestNTStatusAndWin32AreDistinctSpaces is the counterpart of the test in
// win32: the same 32-bit value carries a different meaning in each table, which
// is why they are separate types.
func TestNTStatusAndWin32AreDistinctSpaces(t *testing.T) {
	collisions := []struct {
		value  uint32
		status nt_status.NT_STATUS
		win32  win32.WIN32_ERROR
	}{
		{2, nt_status.NT_STATUS_WAIT_2, win32.ERROR_FILE_NOT_FOUND},
		{3, nt_status.NT_STATUS_WAIT_3, win32.ERROR_PATH_NOT_FOUND},
		{0x103, nt_status.NT_STATUS_PENDING, win32.ERROR_NO_MORE_ITEMS},
	}

	for _, c := range collisions {
		if uint32(c.status) != c.value || uint32(c.win32) != c.value {
			t.Errorf("0x%08X: NTSTATUS is 0x%08X and win32 code is 0x%08X, want both to be the value under test",
				c.value, uint32(c.status), uint32(c.win32))
			continue
		}
		if c.status.Name() == c.win32.Name() {
			t.Errorf("0x%08X resolves to %q in both tables; the two spaces are expected to disagree",
				c.value, c.status.Name())
		}
	}
}

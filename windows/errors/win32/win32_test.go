package win32_test

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// undefined is above the highest code the specification defines (0x00003BC3) and
// has NTSTATUS-shaped severity bits, so it stands for a value that must not be
// found in the table.
const undefined win32.WIN32_ERROR = 0xDEADBEEF

func TestString(t *testing.T) {
	tests := []struct {
		code win32.WIN32_ERROR
		want string
	}{
		{win32.ERROR_SUCCESS, "ERROR_SUCCESS"},
		{win32.ERROR_FILE_NOT_FOUND, "ERROR_FILE_NOT_FOUND"},
		{win32.ERROR_ACCESS_DENIED, "ERROR_ACCESS_DENIED"},
		{win32.ERROR_MORE_DATA, "ERROR_MORE_DATA"},
		{win32.ERROR_NO_MORE_ITEMS, "ERROR_NO_MORE_ITEMS"},
		{undefined, "0xdeadbeef"},
	}

	for _, tt := range tests {
		if got := tt.code.String(); got != tt.want {
			t.Errorf("WIN32_ERROR(0x%08x).String() = %q, want %q", uint32(tt.code), got, tt.want)
		}
	}
}

func TestCodeValuesMatchTheSpecification(t *testing.T) {
	tests := []struct {
		name string
		code win32.WIN32_ERROR
		want uint32
	}{
		{"ERROR_SUCCESS", win32.ERROR_SUCCESS, 0},
		{"NERR_Success", win32.NERR_Success, 0},
		{"ERROR_FILE_NOT_FOUND", win32.ERROR_FILE_NOT_FOUND, 2},
		{"ERROR_ACCESS_DENIED", win32.ERROR_ACCESS_DENIED, 5},
		{"ERROR_NOT_SUPPORTED", win32.ERROR_NOT_SUPPORTED, 50},
		{"ERROR_INVALID_PARAMETER", win32.ERROR_INVALID_PARAMETER, 87},
		{"ERROR_MORE_DATA", win32.ERROR_MORE_DATA, 234},
		{"ERROR_NO_MORE_ITEMS", win32.ERROR_NO_MORE_ITEMS, 259},
		{"ERROR_DS_DRA_ACCESS_DENIED", win32.ERROR_DS_DRA_ACCESS_DENIED, 8453},
	}

	for _, tt := range tests {
		if uint32(tt.code) != tt.want {
			t.Errorf("%s = 0x%08X, want 0x%08X", tt.name, uint32(tt.code), tt.want)
		}
	}
}

func TestNameAndDescription(t *testing.T) {
	if got, want := win32.ERROR_ACCESS_DENIED.Name(), "ERROR_ACCESS_DENIED"; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
	if got, want := win32.ERROR_ACCESS_DENIED.Description(), "Access is denied."; got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}

	// A code carrying an alias reports the canonical name the specification
	// lists first, not the alias.
	if got, want := win32.NERR_Success.Name(), "ERROR_SUCCESS"; got != want {
		t.Errorf("NERR_Success.Name() = %q, want %q", got, want)
	}

	if got := undefined.Name(); got != "" {
		t.Errorf("an undefined code reported the name %q, want the empty string", got)
	}
	if got := undefined.Description(); got != "" {
		t.Errorf("an undefined code reported the description %q, want the empty string", got)
	}
}

func TestIsSuccess(t *testing.T) {
	if !win32.ERROR_SUCCESS.IsSuccess() {
		t.Error("ERROR_SUCCESS did not report success")
	}
	// Unlike an NTSTATUS, a Win32 code has no severity field, so every non-zero
	// value is a failure — including one that reads as informational there.
	for _, code := range []win32.WIN32_ERROR{win32.ERROR_FILE_NOT_FOUND, win32.ERROR_MORE_DATA, win32.ERROR_NO_MORE_ITEMS, undefined} {
		if code.IsSuccess() {
			t.Errorf("0x%08x reported success", uint32(code))
		}
	}
}

func TestError(t *testing.T) {
	if err := win32.ERROR_SUCCESS.Error(); err != nil {
		t.Errorf("ERROR_SUCCESS.Error() = %v, want nil", err)
	}

	err := win32.ERROR_ACCESS_DENIED.Error()
	if err == nil {
		t.Fatal("ERROR_ACCESS_DENIED.Error() returned nil")
	}
	for _, want := range []string{"0x00000005", "ERROR_ACCESS_DENIED", "Access is denied."} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Error() = %q, want it to contain %q", err.Error(), want)
		}
	}

	// A non-zero code the specification does not define is still a failure;
	// reporting nil for it would read as success.
	undefinedErr := undefined.Error()
	if undefinedErr == nil {
		t.Fatal("an undefined non-zero code reported nil, want an error")
	}
	if !strings.Contains(undefinedErr.Error(), "0xdeadbeef") {
		t.Errorf("Error() = %q, want it to contain the hexadecimal code", undefinedErr.Error())
	}
}

func TestLookup(t *testing.T) {
	entry, defined := win32.Lookup(win32.ERROR_INVALID_PARAMETER)
	if !defined {
		t.Fatal("ERROR_INVALID_PARAMETER is not in the table")
	}
	if entry.Name != "ERROR_INVALID_PARAMETER" {
		t.Errorf("entry.Name = %q, want %q", entry.Name, "ERROR_INVALID_PARAMETER")
	}
	if entry.Description == "" {
		t.Error("entry.Description is empty")
	}

	if _, defined := win32.Lookup(undefined); defined {
		t.Error("an undefined code was reported as defined")
	}
}

func TestFromName(t *testing.T) {
	tests := []struct {
		name string
		want win32.WIN32_ERROR
	}{
		{"ERROR_SUCCESS", win32.ERROR_SUCCESS},
		{"NERR_Success", win32.NERR_Success},
		{"ERROR_ACCESS_DENIED", win32.ERROR_ACCESS_DENIED},
		{"WAIT_TIMEOUT", win32.WAIT_TIMEOUT},
		{"RPC_S_INVALID_BINDING", win32.RPC_S_INVALID_BINDING},
		{"DNS_ERROR_RCODE_NAME_ERROR", win32.DNS_ERROR_RCODE_NAME_ERROR},
	}

	for _, tt := range tests {
		got, defined := win32.FromName(tt.name)
		if !defined {
			t.Errorf("FromName(%q) reported the name as undefined", tt.name)
			continue
		}
		if got != tt.want {
			t.Errorf("FromName(%q) = 0x%08X, want 0x%08X", tt.name, uint32(got), uint32(tt.want))
		}
	}

	for _, name := range []string{"", "ERROR_NOT_A_REAL_CODE", "error_access_denied", "STATUS_ACCESS_DENIED"} {
		if code, defined := win32.FromName(name); defined {
			t.Errorf("FromName(%q) resolved to 0x%08X, want undefined", name, uint32(code))
		}
	}
}

// TestWin32AndNTStatusAreDistinctSpaces pins the reason the two tables are
// separate types: the same 32-bit value means different things in each, so a
// value from one is not usable where the other is expected.
func TestWin32AndNTStatusAreDistinctSpaces(t *testing.T) {
	collisions := []struct {
		value    uint32
		win32    win32.WIN32_ERROR
		ntStatus nt_status.NT_STATUS
	}{
		{2, win32.ERROR_FILE_NOT_FOUND, nt_status.NT_STATUS_WAIT_2},
		{3, win32.ERROR_PATH_NOT_FOUND, nt_status.NT_STATUS_WAIT_3},
		{0x103, win32.ERROR_NO_MORE_ITEMS, nt_status.NT_STATUS_PENDING},
	}

	for _, c := range collisions {
		if uint32(c.win32) != c.value || uint32(c.ntStatus) != c.value {
			t.Errorf("0x%08X: win32 code is 0x%08X and NTSTATUS is 0x%08X, want both to be the value under test",
				c.value, uint32(c.win32), uint32(c.ntStatus))
			continue
		}
		if c.win32.Name() == c.ntStatus.String() {
			t.Errorf("0x%08X resolves to %q in both tables; the two spaces are expected to disagree",
				c.value, c.win32.Name())
		}
	}
}

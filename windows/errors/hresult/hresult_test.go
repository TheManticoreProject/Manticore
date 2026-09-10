package hresult_test

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// undefined is a failure in a facility the specification does not populate, so
// neither the table nor the FACILITY_WIN32 derivation can name it.
const undefined hresult.HRESULT = 0x8DEA0001

func TestString(t *testing.T) {
	tests := []struct {
		value hresult.HRESULT
		want  string
	}{
		{hresult.S_OK, "S_OK"},
		{hresult.S_FALSE, "S_FALSE"},
		{hresult.E_FAIL, "E_FAIL"},
		{hresult.E_ACCESSDENIED, "E_ACCESSDENIED"},
		{hresult.STG_S_CONVERTED, "STG_S_CONVERTED"},
		// A FACILITY_WIN32 failure the specification does not name resolves
		// through the Win32 table it wraps, and says so rather than passing the
		// Win32 name off as an HRESULT name.
		{hresult.FromWin32(win32.ERROR_INVALID_NAME), "HRESULT_FROM_WIN32(ERROR_INVALID_NAME)"},
		{hresult.FromWin32(win32.ERROR_FILE_NOT_FOUND), "HRESULT_FROM_WIN32(ERROR_FILE_NOT_FOUND)"},
		{undefined, "0x8dea0001"},
	}

	for _, tt := range tests {
		if got := tt.value.String(); got != tt.want {
			t.Errorf("HRESULT(0x%08x).String() = %q, want %q", uint32(tt.value), got, tt.want)
		}
	}
}

// TestSpecificationNameWinsOverDerivation pins the precedence in lookup: the
// five FACILITY_WIN32 values [MS-ERREF] 2.1.1 names itself report that name,
// not the HRESULT_FROM_WIN32 form.
func TestSpecificationNameWinsOverDerivation(t *testing.T) {
	cases := map[hresult.HRESULT]string{
		0x80070005: "E_ACCESSDENIED",
		0x8007000E: "E_OUTOFMEMORY",
		0x80070057: "E_INVALIDARG",
		0x80070032: "ERROR_NOT_SUPPORTED",
		0x80070070: "ERROR_DISK_FULL",
	}

	for value, want := range cases {
		if got := value.String(); got != want {
			t.Errorf("0x%08X renders as %q, want the specification's own name %q", uint32(value), got, want)
		}
		// Each is still a FACILITY_WIN32 value that unwraps.
		if _, wraps := value.ToWin32(); !wraps {
			t.Errorf("0x%08X does not report as wrapping a Win32 code", uint32(value))
		}
	}
}

// TestIsSuccessIsARange is the property that separates HRESULT from the other
// two tables: success is the whole severity-clear half of the space, not the
// single value zero.
func TestIsSuccessIsARange(t *testing.T) {
	for _, value := range []hresult.HRESULT{hresult.S_OK, hresult.S_FALSE, hresult.STG_S_CONVERTED, hresult.XACT_S_READONLY} {
		if !value.IsSuccess() {
			t.Errorf("0x%08X did not report success", uint32(value))
		}
		if err := value.Error(); err != nil {
			t.Errorf("0x%08X reported the error %v, want nil for a success-severity value", uint32(value), err)
		}
	}

	for _, value := range []hresult.HRESULT{hresult.E_FAIL, hresult.E_ACCESSDENIED, hresult.E_UNEXPECTED, undefined} {
		if value.IsSuccess() {
			t.Errorf("0x%08X reported success", uint32(value))
		}
		if value.Error() == nil {
			t.Errorf("0x%08X reported nil, want an error", uint32(value))
		}
	}
}

func TestError(t *testing.T) {
	err := hresult.E_ACCESSDENIED.Error()
	if err == nil {
		t.Fatal("E_ACCESSDENIED.Error() returned nil")
	}
	for _, want := range []string{"0x80070005", "E_ACCESSDENIED"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Error() = %q, want it to contain %q", err.Error(), want)
		}
	}

	undefinedErr := undefined.Error()
	if undefinedErr == nil {
		t.Fatal("an undefined failure reported nil, want an error")
	}
	if !strings.Contains(undefinedErr.Error(), "0x8dea0001") {
		t.Errorf("Error() = %q, want it to contain the hexadecimal value", undefinedErr.Error())
	}
}

func TestNameAndDescription(t *testing.T) {
	if got, want := hresult.E_FAIL.Name(), "E_FAIL"; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
	if hresult.E_FAIL.Description() == "" {
		t.Error("E_FAIL has no description")
	}
	// A derived value borrows the wrapped Win32 code's description.
	wrapped := hresult.FromWin32(win32.ERROR_FILE_NOT_FOUND)
	if got, want := wrapped.Description(), win32.ERROR_FILE_NOT_FOUND.Description(); got != want {
		t.Errorf("a derived value's description = %q, want the wrapped code's %q", got, want)
	}
	if got := undefined.Name(); got != "" {
		t.Errorf("an undefined value reported the name %q, want the empty string", got)
	}
}

func TestLookupAndFromName(t *testing.T) {
	entry, defined := hresult.Lookup(hresult.E_INVALIDARG)
	if !defined || entry.Name != "E_INVALIDARG" {
		t.Errorf("Lookup(E_INVALIDARG) = %+v, %v", entry, defined)
	}
	if _, defined := hresult.Lookup(undefined); defined {
		t.Error("an undefined value was reported as defined")
	}

	for name, want := range map[string]hresult.HRESULT{
		"S_OK":           hresult.S_OK,
		"S_FALSE":        hresult.S_FALSE,
		"E_FAIL":         hresult.E_FAIL,
		"E_ACCESSDENIED": hresult.E_ACCESSDENIED,
	} {
		got, defined := hresult.FromName(name)
		if !defined || got != want {
			t.Errorf("FromName(%q) = 0x%08X, %v; want 0x%08X, true", name, uint32(got), defined, uint32(want))
		}
	}

	// The derived form is not a name the table resolves; FromWin32 builds those.
	for _, name := range []string{"", "HRESULT_FROM_WIN32(ERROR_FILE_NOT_FOUND)", "ERROR_FILE_NOT_FOUND", "e_fail"} {
		if value, defined := hresult.FromName(name); defined {
			t.Errorf("FromName(%q) resolved to 0x%08X, want undefined", name, uint32(value))
		}
	}
}

// TestFromWin32MatchesTheMacro checks the arithmetic of [MS-ERREF] 2.1.2
// directly, including the zero case the macro maps through unchanged.
func TestFromWin32MatchesTheMacro(t *testing.T) {
	if got := hresult.FromWin32(win32.ERROR_SUCCESS); got != hresult.S_OK {
		t.Errorf("FromWin32(ERROR_SUCCESS) = 0x%08X, want S_OK", uint32(got))
	}

	for _, code := range []win32.WIN32_ERROR{
		win32.ERROR_FILE_NOT_FOUND, win32.ERROR_ACCESS_DENIED,
		win32.ERROR_INVALID_PARAMETER, win32.ERROR_MORE_DATA,
		win32.ERROR_DS_DRA_ACCESS_DENIED,
	} {
		value := hresult.FromWin32(code)
		if want := hresult.HRESULT(0x80070000 | uint32(code)&0xFFFF); value != want {
			t.Errorf("FromWin32(0x%08X) = 0x%08X, want 0x%08X", uint32(code), uint32(value), uint32(want))
		}
		if value.IsSuccess() {
			t.Errorf("FromWin32(0x%08X) reported success", uint32(code))
		}
		back, wraps := value.ToWin32()
		if !wraps || back != code {
			t.Errorf("0x%08X.ToWin32() = 0x%08X, %v; want 0x%08X, true", uint32(value), uint32(back), wraps, uint32(code))
		}
	}
}

func TestToWin32OnlyUnwrapsFacilityWin32(t *testing.T) {
	if code, wraps := hresult.S_OK.ToWin32(); !wraps || code != win32.ERROR_SUCCESS {
		t.Errorf("S_OK.ToWin32() = 0x%08X, %v; want ERROR_SUCCESS, true", uint32(code), wraps)
	}
	// E_FAIL is FACILITY_NULL, XACT_S_READONLY is a success — neither wraps.
	for _, value := range []hresult.HRESULT{hresult.E_FAIL, hresult.E_UNEXPECTED, hresult.XACT_S_READONLY, hresult.S_FALSE} {
		if code, wraps := value.ToWin32(); wraps {
			t.Errorf("0x%08X.ToWin32() reported it wraps 0x%08X, want false", uint32(value), uint32(code))
		}
	}
}

// TestTheThreeTablesDisagree pins the reason windows/errors holds three separate
// types. The same 32-bit value carries a different meaning in each, and the
// success rule itself differs, so no caller can treat one as a stand-in for
// another.
func TestTheThreeTablesDisagree(t *testing.T) {
	// 0x00000001 is a success in the HRESULT space and a failure in both others.
	if !hresult.S_FALSE.IsSuccess() {
		t.Error("S_FALSE is not a success in the HRESULT space")
	}
	if got, want := win32.WIN32_ERROR(1).Name(), "ERROR_INVALID_FUNCTION"; got != want {
		t.Errorf("win32 names 0x00000001 %q, want %q", got, want)
	}
	if win32.WIN32_ERROR(1).Error() == nil {
		t.Error("0x00000001 is a success in the win32 space; it should be a failure")
	}
	if nt_status.NT_STATUS(1).Error() == nil {
		t.Error("0x00000001 is a success in the NTSTATUS space; it should be a failure")
	}

	// Success is a range here and a single value in the other two.
	if hresult.STG_S_CONVERTED.Error() != nil {
		t.Error("a success-severity HRESULT reported an error")
	}
	if win32.WIN32_ERROR(0x00030200).Error() == nil {
		t.Error("win32 treats 0x00030200 as a success; only zero should succeed there")
	}
}

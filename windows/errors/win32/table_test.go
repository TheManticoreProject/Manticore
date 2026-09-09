package win32

import "testing"

// codeFloor guards against a regeneration that silently drops rows. The
// specification defined 2703 codes when the table was generated; the floor is
// set just below so that growth is free and loss is not.
const codeFloor = 2700

func TestTableIsComplete(t *testing.T) {
	if len(table) < codeFloor {
		t.Errorf("the table holds %d codes, want at least %d", len(table), codeFloor)
	}
	if len(nameToCode) < len(table) {
		t.Errorf("nameToCode holds %d names for %d codes; every code needs at least its canonical name",
			len(nameToCode), len(table))
	}

	for code, entry := range table {
		if entry.Name == "" {
			t.Errorf("0x%08X has no name", uint32(code))
		}
		if entry.Description == "" {
			t.Errorf("%s (0x%08X) has no description", entry.Name, uint32(code))
		}
	}
}

// TestCanonicalNamesResolveBackToTheirCode is the consistency check between the
// two generated maps: a name read out of one must resolve through the other to
// the code it was read from.
func TestCanonicalNamesResolveBackToTheirCode(t *testing.T) {
	for code, entry := range table {
		resolved, defined := nameToCode[entry.Name]
		if !defined {
			t.Errorf("%s (0x%08X) is the canonical name of a code but is absent from nameToCode",
				entry.Name, uint32(code))
			continue
		}
		if resolved != code {
			t.Errorf("%s is the canonical name of 0x%08X but nameToCode resolves it to 0x%08X",
				entry.Name, uint32(code), uint32(resolved))
		}
	}
}

// TestEveryNameNamesADefinedCode covers the other direction, including aliases:
// no name may resolve to a code with no entry.
func TestEveryNameNamesADefinedCode(t *testing.T) {
	for name, code := range nameToCode {
		if _, defined := table[code]; !defined {
			t.Errorf("%s resolves to 0x%08X, which has no entry in the table", name, uint32(code))
		}
	}
}

// TestCodesAreFlatWin32Values pins the property that distinguishes this table
// from NTSTATUS: a Win32 error code is a flat integer with no severity or
// facility bits. A value with the high bits set would mean an NTSTATUS had been
// extracted into this table.
func TestCodesAreFlatWin32Values(t *testing.T) {
	for code := range table {
		if uint32(code) > 0xFFFF {
			t.Errorf("%s is 0x%08X, which is outside the flat Win32 range and looks like an NTSTATUS",
				table[code].Name, uint32(code))
		}
	}
}

// TestSuccessIsTheOnlyZeroCode checks the canonical name chosen for the one code
// the specification gives two names, and that success is a single value rather
// than a range as it is for NTSTATUS.
func TestSuccessIsTheOnlyZeroCode(t *testing.T) {
	if got, want := table[ERROR_SUCCESS].Name, "ERROR_SUCCESS"; got != want {
		t.Errorf("the canonical name of 0x00000000 is %q, want %q", got, want)
	}
	if ERROR_SUCCESS != 0 || NERR_Success != 0 {
		t.Errorf("ERROR_SUCCESS = 0x%08X and NERR_Success = 0x%08X, want both to be zero",
			uint32(ERROR_SUCCESS), uint32(NERR_Success))
	}

	for code := range table {
		if code.IsSuccess() && code != ERROR_SUCCESS {
			t.Errorf("%s (0x%08X) reports success", table[code].Name, uint32(code))
		}
	}
}

// TestErrorTextIsStable pins the rendered form of an error, since it reaches
// logs and comparisons by string in calling code.
func TestErrorTextIsStable(t *testing.T) {
	const want = "WIN32_ERROR(0x00000005): ERROR_ACCESS_DENIED: Access is denied."
	if got := ERROR_ACCESS_DENIED.Error().Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	if err := ERROR_SUCCESS.Error(); err != nil {
		t.Errorf("ERROR_SUCCESS.Error() = %v, want nil", err)
	}
}

package hresult

import "testing"

// codeFloor guards against a regeneration that silently drops rows. [MS-ERREF]
// 2.1.1 defined 2928 values when the table was generated; the floor is set just
// below so that growth is free and loss is not.
const codeFloor = 2900

func TestTableIsComplete(t *testing.T) {
	if len(table) < codeFloor {
		t.Errorf("the table holds %d values, want at least %d", len(table), codeFloor)
	}
	if len(nameToCode) < len(table) {
		t.Errorf("nameToCode holds %d names for %d values", len(nameToCode), len(table))
	}

	for value, entry := range table {
		if entry.Name == "" {
			t.Errorf("0x%08X has no name", uint32(value))
		}
		if entry.Description == "" {
			t.Errorf("%s (0x%08X) has no description", entry.Name, uint32(value))
		}
	}
}

func TestNamesResolveBackToTheirValue(t *testing.T) {
	for value, entry := range table {
		resolved, defined := nameToCode[entry.Name]
		if !defined {
			t.Errorf("%s (0x%08X) is absent from nameToCode", entry.Name, uint32(value))
			continue
		}
		if resolved != value {
			t.Errorf("%s names 0x%08X but resolves to 0x%08X", entry.Name, uint32(value), uint32(resolved))
		}
	}
	for name, value := range nameToCode {
		if _, defined := table[value]; !defined {
			t.Errorf("%s resolves to 0x%08X, which has no entry", name, uint32(value))
		}
	}
}

// TestWellKnownValuesAreAbsentFromTheSpecificationTable records why
// wellknown.go exists: [MS-ERREF] 2.1.1 starts at 0x00030200 and lists neither
// S_OK nor S_FALSE, so the generated table cannot name the two most common
// HRESULT values.
func TestWellKnownValuesAreAbsentFromTheSpecificationTable(t *testing.T) {
	for _, value := range []HRESULT{S_OK, S_FALSE} {
		if _, inTable := table[value]; inTable {
			t.Errorf("0x%08X is now in the generated table; wellknown.go should drop it", uint32(value))
		}
		if _, defined := wellKnown[value]; !defined {
			t.Errorf("0x%08X is in neither the generated table nor wellKnown", uint32(value))
		}
		if _, defined := lookup(value); !defined {
			t.Errorf("0x%08X does not resolve through lookup", uint32(value))
		}
	}

	lowest := ^uint32(0)
	for value := range table {
		if uint32(value) < lowest {
			lowest = uint32(value)
		}
	}
	if lowest <= uint32(S_FALSE) {
		t.Errorf("the generated table's lowest value is 0x%08X, which reaches the well-known range", lowest)
	}
}

// TestTableSpansBothSeverities pins that the table carries success-severity
// values as well as failures, which is what makes IsSuccess a range test rather
// than an equality test.
func TestTableSpansBothSeverities(t *testing.T) {
	var successes, failures int
	for value := range table {
		if value.IsSuccess() {
			successes++
		} else {
			failures++
		}
	}
	if successes == 0 {
		t.Error("the table holds no success-severity value")
	}
	if failures == 0 {
		t.Error("the table holds no failure value")
	}
	if successes >= failures {
		t.Errorf("severity split looks wrong: %d successes, %d failures", successes, failures)
	}
}

// TestFacilityWin32CoverageComesFromDerivation records that the specification
// names only a handful of FACILITY_WIN32 values itself, and that the rest are
// reached by computation rather than transcription — the reason no synthetic
// rows were generated for the 2703 Win32 codes.
func TestFacilityWin32CoverageComesFromDerivation(t *testing.T) {
	var named int
	for value := range table {
		if !value.IsSuccess() && value.facility() == facilityWin32 {
			named++
		}
	}
	if named == 0 {
		t.Fatal("the table names no FACILITY_WIN32 value; the precedence test is then meaningless")
	}
	if named > 100 {
		t.Errorf("the table names %d FACILITY_WIN32 values; the derivation may now be redundant", named)
	}
}

func TestErrorTextIsStable(t *testing.T) {
	const want = "HRESULT(0x80004005): E_FAIL: Unspecified error."
	if got := E_FAIL.Error().Error(); got != want {
		t.Errorf("Error() =\n  %q\nwant\n  %q", got, want)
	}
	if err := S_OK.Error(); err != nil {
		t.Errorf("S_OK.Error() = %v, want nil", err)
	}
}

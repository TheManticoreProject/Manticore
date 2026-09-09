package nt_status

import "testing"

// codeFloor guards against a regeneration that silently drops rows. [MS-ERREF]
// 2.3.1 defined 1796 codes when the table was generated; the floor is set just
// below so that growth is free and loss is not.
const codeFloor = 1790

func TestTableIsComplete(t *testing.T) {
	if len(table) < codeFloor {
		t.Errorf("the table holds %d codes, want at least %d", len(table), codeFloor)
	}
	if len(nameToCode) < len(table) {
		t.Errorf("nameToCode holds %d names for %d codes; every code needs at least its canonical name",
			len(nameToCode), len(table))
	}

	for status, entry := range table {
		if entry.Name == "" {
			t.Errorf("0x%08X has no name", uint32(status))
		}
		if entry.Description == "" {
			t.Errorf("%s (0x%08X) has no description", entry.Name, uint32(status))
		}
	}
}

// TestCanonicalNamesResolveBackToTheirStatus is the consistency check between
// the two generated maps: a name read out of one must resolve through the other
// to the status it was read from.
func TestCanonicalNamesResolveBackToTheirStatus(t *testing.T) {
	for status, entry := range table {
		resolved, defined := nameToCode[entry.Name]
		if !defined {
			t.Errorf("%s (0x%08X) is the canonical name of a status but is absent from nameToCode",
				entry.Name, uint32(status))
			continue
		}
		if resolved != status {
			t.Errorf("%s is the canonical name of 0x%08X but nameToCode resolves it to 0x%08X",
				entry.Name, uint32(status), uint32(resolved))
		}
	}
}

// TestEveryNameNamesADefinedStatus covers the other direction, including the
// specification's own spelling of each name: none may resolve to a status with
// no entry.
func TestEveryNameNamesADefinedStatus(t *testing.T) {
	for name, status := range nameToCode {
		if _, defined := table[status]; !defined {
			t.Errorf("%s resolves to 0x%08X, which has no entry in the table", name, uint32(status))
		}
	}
}

// TestSpecificationNamesAlsoResolve asserts the generator emitted the [MS-ERREF]
// spelling alongside the Go constant name, so a status copied out of the
// specification resolves without translation.
func TestSpecificationNamesAlsoResolve(t *testing.T) {
	cases := map[string]NT_STATUS{
		"STATUS_SUCCESS":       NT_STATUS_SUCCESS,
		"STATUS_ACCESS_DENIED": NT_STATUS_ACCESS_DENIED,
		"STATUS_PENDING":       NT_STATUS_PENDING,
		// A name the specification does not prefix with STATUS_ keeps its own
		// spelling on both sides.
		"DBG_EXCEPTION_HANDLED": NT_STATUS_DBG_EXCEPTION_HANDLED,
		"RPC_NT_BAD_STUB_DATA":  NT_STATUS_RPC_NT_BAD_STUB_DATA,
	}

	for name, want := range cases {
		got, defined := nameToCode[name]
		if !defined {
			t.Errorf("the specification name %s does not resolve", name)
			continue
		}
		if got != want {
			t.Errorf("%s resolves to 0x%08X, want 0x%08X", name, uint32(got), uint32(want))
		}
	}
}

// TestTableSpansTheSeverityField pins the property that distinguishes this table
// from the Win32 one: an NTSTATUS is a bitfield whose top two bits carry
// severity, so the table must hold values in each severity and well above the
// flat range a Win32 error code occupies. A table confined to small integers
// would mean the Win32 codes had been generated into this package.
func TestTableSpansTheSeverityField(t *testing.T) {
	counts := map[uint32]int{}
	for status := range table {
		counts[uint32(status)>>30]++
	}

	for severity, label := range map[uint32]string{0: "success", 1: "informational", 2: "warning", 3: "error"} {
		if counts[severity] == 0 {
			t.Errorf("the table holds no %s status", label)
		}
	}
	if counts[3] == 0 || counts[3] < counts[0] {
		t.Errorf("severity counts look wrong: %v", counts)
	}
}

// TestSuccessIsTheZeroStatus checks the canonical name chosen for 0x00000000,
// which the specification names both STATUS_SUCCESS and STATUS_WAIT_0.
func TestSuccessIsTheZeroStatus(t *testing.T) {
	if got, want := table[NT_STATUS_SUCCESS].Name, "NT_STATUS_SUCCESS"; got != want {
		t.Errorf("the canonical name of 0x00000000 is %q, want %q", got, want)
	}
	if NT_STATUS_SUCCESS != 0 || NT_STATUS_WAIT_0 != 0 {
		t.Errorf("NT_STATUS_SUCCESS = 0x%08X and NT_STATUS_WAIT_0 = 0x%08X, want both to be zero",
			uint32(NT_STATUS_SUCCESS), uint32(NT_STATUS_WAIT_0))
	}
	// The displaced alias still resolves by name.
	if got, defined := nameToCode["NT_STATUS_WAIT_0"]; !defined || got != NT_STATUS_SUCCESS {
		t.Errorf("NT_STATUS_WAIT_0 resolves to 0x%08X, %v; want 0x00000000, true", uint32(got), defined)
	}
}

// TestErrorTextIsStable pins the rendered form of an error, since it reaches
// logs and comparisons by string in calling code.
func TestErrorTextIsStable(t *testing.T) {
	const want = "NT_STATUS(0xc0000022): NT_STATUS_ACCESS_DENIED: {Access Denied} A process has requested access to an object but has not been granted those access rights."
	if got := NT_STATUS_ACCESS_DENIED.Error().Error(); got != want {
		t.Errorf("Error() =\n  %q\nwant\n  %q", got, want)
	}
	if err := NT_STATUS_SUCCESS.Error(); err != nil {
		t.Errorf("NT_STATUS_SUCCESS.Error() = %v, want nil", err)
	}
}

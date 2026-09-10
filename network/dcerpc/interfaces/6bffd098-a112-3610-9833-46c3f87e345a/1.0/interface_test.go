package rpcinterface_6bffd098a1123610983346c3f87e345a_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// TestSyntaxID verifies the abstract syntax identifier matches [MS-WKST] 1.9:
// 6BFFD098-A112-3610-9833-46C3F87E345A, version 1.0.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	want := guid.GUID{A: 0x6bffd098, B: 0xa112, C: 0x3610, D: 0x9833, E: 0x46c3f87e345a}
	if s.UUID != want {
		t.Errorf("UUID = %+v, want %+v (6bffd098-a112-3610-9833-46c3f87e345a)", s.UUID, want)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName pins the transport endpoint.
func TestPipeName(t *testing.T) {
	if PipeName != `\wkssvc` {
		t.Errorf("PipeName = %q, want %q", PipeName, `\wkssvc`)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are consistent inverses and
// that the on-the-wire opnums match [MS-WKST] 3.2.4 (the ten NotUsedOnWire opnums absent).
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 28 {
		t.Errorf("OpnumToName has %d entries, want 28 on-the-wire methods", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d (ok=%v), want %d", name, got, ok, op)
		}
	}

	// A NotUsedOnWire opnum must not be present.
	for _, gap := range []uint16{3, 4, 12, 14, 15, 16, 17, 18, 19, 21} {
		if name, ok := OpnumToName[gap]; ok {
			t.Errorf("opnum %d is NotUsedOnWire but mapped to %q", gap, name)
		}
	}

	// Spot-check a few concrete opnums.
	for name, want := range map[string]uint16{
		"NetrWkstaGetInfo":            0,
		"NetrUseEnum":                 11,
		"NetrGetJoinInformation":      20,
		"NetrJoinDomain2":             22,
		"NetrSetPrimaryComputerName2": 37,
	} {
		if got := NameToOpnum[name]; got != want {
			t.Errorf("%s opnum = %d, want %d", name, got, want)
		}
	}
}

// TestStatusCodesResolveThroughWin32 checks that the NET_API_STATUS codes this
// interface documents in [MS-WKST] 3.2.4 resolve through the shared [MS-ERREF]
// 2.2 table under their specification names, that a code outside the subset the
// interface used to declare now renders by name instead of as undecoded hex, and
// that a value the specification does not define still falls back to hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	cases := map[win32.WIN32_ERROR]string{
		// 0x00000000 carries both names in the specification; the table renders
		// the ERROR_* one, and win32.NERR_Success is the same value.
		win32.NERR_Success:              "ERROR_SUCCESS",
		win32.ERROR_ACCESS_DENIED:       "ERROR_ACCESS_DENIED",
		win32.ERROR_MORE_DATA:           "ERROR_MORE_DATA",
		win32.ERROR_INVALID_DOMAIN_ROLE: "ERROR_INVALID_DOMAIN_ROLE",
		win32.NERR_SetupAlreadyJoined:   "NERR_SetupAlreadyJoined",
		win32.NERR_WkstaNotStarted:      "NERR_WkstaNotStarted",

		// Outside the subset that used to live in interface.go: these rendered as
		// "0x000004ba" and "0x00000102" before, and now carry their names.
		win32.WIN32_ERROR(0x000004BA): "ERROR_INVALID_COMPUTERNAME",
		win32.WIN32_ERROR(0x00000102): "WAIT_TIMEOUT",

		// Not defined by the specification at all: still hex.
		win32.WIN32_ERROR(0x12345678): "0x12345678",
	}
	for code, want := range cases {
		if got := code.String(); got != want {
			t.Errorf("WIN32_ERROR(0x%08x).String() = %q, want %q", uint32(code), got, want)
		}
	}
}

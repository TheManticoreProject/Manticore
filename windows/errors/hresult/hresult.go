// Package hresult is the HRESULT table of [MS-ERREF] section 2.1, the status
// values returned by COM and by the DCOM-adjacent RPC interfaces that report
// failure that way.
//
// An HRESULT is a bitfield: a severity bit, a customer flag, reserved bits, an
// 11-bit facility and a 16-bit code. Two consequences separate it from the other
// two tables in windows/errors:
//
// Success is a range, not a value. Bit 31 clear means success, so S_OK, S_FALSE
// and the 159 other success-severity values in the specification are all
// successes, and [HRESULT.Error] reports nil for every one of them. In
// [github.com/TheManticoreProject/Manticore/windows/errors/win32] and
// [github.com/TheManticoreProject/Manticore/windows/errors/nt_status] success is
// the single value zero.
//
// Part of the space is a mechanical wrapping of another table. FACILITY_WIN32
// (7) means the low 16 bits carry a Win32 error code, per the macro in
// [MS-ERREF] 2.1.2, so 0x8007XXXX is HRESULT_FROM_WIN32 of the Win32 code
// 0x0000XXXX. The specification's own table names only five of those, so the
// rest are resolved by computation rather than by transcription: see
// [FromWin32], [HRESULT.ToWin32] and the fallback in lookup. That is the one
// place where two of these tables genuinely relate by arithmetic; NTSTATUS and
// Win32 relate only by a many-to-one lookup, which is why no conversion is
// offered between those two.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-erref/705fb797-2175-4a90-b5a3-3918024b10b8
package hresult

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

//go:generate go run ../errgen -in ../errgen/ms-erref-2.1.1-hresult.tsv -out . -package hresult -type HRESULT -source https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-erref/705fb797-2175-4a90-b5a3-3918024b10b8

// HRESULT is a COM status value.
type HRESULT uint32

// Entry is what the specification records for one value.
type Entry struct {
	// Name is the value's canonical symbolic name, such as "E_ACCESSDENIED".
	Name string
	// Description is the description the specification gives, verbatim. A
	// percentage sign followed by alphanumeric characters — "%1", "%hs" — marks
	// a variable that Windows replaces with text when the value is returned.
	Description string
}

// facilityWin32 is FACILITY_WIN32, the facility whose low 16 bits carry a Win32
// error code ([MS-ERREF] 2.1.2).
const facilityWin32 = 0x0007

// facility returns the value's 11-bit facility field.
func (h HRESULT) facility() uint16 {
	return uint16((h >> 16) & 0x07FF)
}

// FromWin32 wraps a Win32 error code as an HRESULT, applying the
// HRESULT_FROM_WIN32 macro of [MS-ERREF] 2.1.2: a positive code becomes
// 0x8007XXXX over its low 16 bits, and zero is carried through as S_OK.
//
// The macro passes a value that is already negative through unchanged. A
// [win32.WIN32_ERROR] is never negative — the Win32 table tops out at
// 0x00003BC3 — so that arm cannot arise here and is not reproduced.
func FromWin32(code win32.WIN32_ERROR) HRESULT {
	if code == win32.ERROR_SUCCESS {
		return S_OK
	}
	return HRESULT(uint32(code)&0x0000FFFF | facilityWin32<<16 | 0x80000000)
}

// ToWin32 returns the Win32 error code an HRESULT wraps, and whether it wraps
// one at all. Only a failure in FACILITY_WIN32 does; S_OK reports
// ERROR_SUCCESS, since the macro maps zero to zero in both directions.
//
// The result is the code the macro would have been given, which is not
// necessarily a code the Win32 table defines: the low 16 bits are carried
// verbatim, so an HRESULT built from a value outside that table round-trips to
// the same value outside it.
func (h HRESULT) ToWin32() (win32.WIN32_ERROR, bool) {
	if h == S_OK {
		return win32.ERROR_SUCCESS, true
	}
	if h.IsSuccess() || h.facility() != facilityWin32 {
		return 0, false
	}
	return win32.WIN32_ERROR(uint32(h) & 0x0000FFFF), true
}

// lookup resolves a value against the specification's table, then the
// well-known values [MS-ERREF] 2.1.1 omits, then — for a FACILITY_WIN32 failure
// the first two do not name — the Win32 table it wraps.
//
// The specification's table wins over the derivation, so 0x80070005 reports
// E_ACCESSDENIED, the name [MS-ERREF] 2.1.1 gives it, rather than the
// ERROR_ACCESS_DENIED it is built from.
func lookup(h HRESULT) (Entry, bool) {
	if entry, defined := table[h]; defined {
		return entry, true
	}
	if entry, defined := wellKnown[h]; defined {
		return entry, true
	}
	if code, wraps := h.ToWin32(); wraps {
		if wrapped, defined := win32.Lookup(code); defined {
			return Entry{
				Name:        "HRESULT_FROM_WIN32(" + wrapped.Name + ")",
				Description: wrapped.Description,
			}, true
		}
	}
	return Entry{}, false
}

// String returns the value's symbolic name, or its hexadecimal value when
// neither the specification nor the Win32 table it may wrap defines one. It
// never returns the empty string, so it is safe to interpolate into a log line
// directly.
//
// A FACILITY_WIN32 value the specification does not name renders as
// HRESULT_FROM_WIN32(ERROR_INVALID_NAME) rather than as the bare Win32 name,
// since the value is an HRESULT and not the code it wraps.
func (h HRESULT) String() string {
	if entry, defined := lookup(h); defined {
		return entry.Name
	}
	return fmt.Sprintf("0x%08x", uint32(h))
}

// Name returns the value's symbolic name, or the empty string when none is
// defined. Use [HRESULT.String] to render a value whose name may be unknown.
func (h HRESULT) Name() string {
	entry, _ := lookup(h)
	return entry.Name
}

// Description returns the description the specification gives for the value, or
// the empty string when it defines none. A value resolved through
// FACILITY_WIN32 reports the description of the Win32 code it wraps.
func (h HRESULT) Description() string {
	entry, _ := lookup(h)
	return entry.Description
}

// IsSuccess reports whether the severity bit is clear, which for an HRESULT
// means success ([MS-ERREF] 2.1).
//
// This is a range, not a single value: S_OK, S_FALSE and every other
// success-severity value satisfy it. That differs from
// [github.com/TheManticoreProject/Manticore/windows/errors/win32] and
// [github.com/TheManticoreProject/Manticore/windows/errors/nt_status], where
// success is zero alone, so a caller switching between the three cannot assume
// one rule.
func (h HRESULT) IsSuccess() bool {
	return h>>31 == 0
}

// Error returns nil for every success-severity value and an error for every
// failure, including one neither the specification nor the wrapped Win32 table
// defines — an undefined failure is still a failure, and reporting nil for it
// would read as success.
func (h HRESULT) Error() error {
	if h.IsSuccess() {
		return nil
	}
	if entry, defined := lookup(h); defined {
		return fmt.Errorf("HRESULT(0x%08x): %s: %s", uint32(h), entry.Name, entry.Description)
	}
	return fmt.Errorf("HRESULT(0x%08x)", uint32(h))
}

// Lookup returns the specification's entry for a value, and whether one is
// defined. A FACILITY_WIN32 value resolves through the Win32 table it wraps.
func Lookup(value HRESULT) (Entry, bool) {
	return lookup(value)
}

// FromName resolves a symbolic name to its value, and reports whether the name
// is defined. The well-known names [MS-ERREF] 2.1.1 omits resolve too, so
// FromName("S_OK") and FromName("E_FAIL") both work.
//
// Names of the HRESULT_FROM_WIN32(...) form are not resolved; build those with
// [FromWin32] from the Win32 constant instead.
func FromName(name string) (HRESULT, bool) {
	if value, defined := nameToCode[name]; defined {
		return value, true
	}
	value, defined := wellKnownNames[name]
	return value, defined
}

// Package win32 is the Win32 system error code table of [MS-ERREF] section 2.2,
// the codes returned as a DWORD by the Win32 API and by the RPC methods that
// report failure that way.
//
// These are not NTSTATUS values, and the two are not interchangeable. An
// NTSTATUS is a bitfield — severity, customer, N, facility and code — whereas a
// Win32 error code is a flat integer, none of them above 0x00003BC3 in the
// current specification. The two spaces also collide by value: 0x00000002 is
// ERROR_FILE_NOT_FOUND here and STATUS_WAIT_2 in
// [github.com/TheManticoreProject/Manticore/windows/errors/nt_status], and
// 0x00000103 is ERROR_NO_MORE_ITEMS here and STATUS_PENDING there. Which table
// applies is a property of the field being read, so the two are separate types
// and a value from one cannot be passed where the other is expected.
//
// Converting between them is a lookup rather than a computation, and no
// conversion is offered here.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-erref/18d8fbe8-a967-4f1c-ae50-99ca8e491d2d
package win32

import "fmt"

//go:generate go run ../errgen -in ../errgen/ms-erref-2.2-win32.tsv -out . -package win32 -type WIN32_ERROR -source https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-erref/18d8fbe8-a967-4f1c-ae50-99ca8e491d2d

// WIN32_ERROR is a Win32 system error code.
type WIN32_ERROR uint32

// Entry is what the specification records for one code.
type Entry struct {
	// Name is the code's canonical symbolic name, such as
	// "ERROR_ACCESS_DENIED".
	Name string
	// Description is the description the specification gives, verbatim. A
	// percentage sign followed by alphanumeric characters — "%1", "%hs" — marks
	// a variable that Windows replaces with text when the code is returned.
	Description string
}

// String returns the code's symbolic name, or its hexadecimal value when the
// specification defines no name for it. It never returns the empty string, so it
// is safe to interpolate into a log line directly.
func (e WIN32_ERROR) String() string {
	if entry, defined := table[e]; defined {
		return entry.Name
	}
	return fmt.Sprintf("0x%08x", uint32(e))
}

// Name returns the code's canonical symbolic name, or the empty string when the
// specification defines no name for it. Use [WIN32_ERROR.String] to render a
// code whose name may be unknown.
func (e WIN32_ERROR) Name() string {
	return table[e].Name
}

// Description returns the description the specification gives for the code, or
// the empty string when it defines none.
func (e WIN32_ERROR) Description() string {
	return table[e].Description
}

// IsSuccess reports whether the code is ERROR_SUCCESS. Unlike an NTSTATUS, a
// Win32 error code carries no severity field, so success is the single value
// zero and every other code is a failure.
func (e WIN32_ERROR) IsSuccess() bool {
	return e == ERROR_SUCCESS
}

// Error returns nil for ERROR_SUCCESS and an error for every other code,
// including one the specification does not define — an undefined non-zero code
// is still a failure, and reporting nil for it would read as success.
func (e WIN32_ERROR) Error() error {
	if e.IsSuccess() {
		return nil
	}
	if entry, defined := table[e]; defined {
		return fmt.Errorf("WIN32_ERROR(0x%08x): %s: %s", uint32(e), entry.Name, entry.Description)
	}
	return fmt.Errorf("WIN32_ERROR(0x%08x)", uint32(e))
}

// Lookup returns the specification's entry for a code, and whether it defines
// one.
func Lookup(code WIN32_ERROR) (Entry, bool) {
	entry, defined := table[code]
	return entry, defined
}

// FromName resolves a symbolic name to its code, and reports whether the name is
// defined. Both a canonical name and an alias resolve, so FromName("NERR_Success")
// and FromName("ERROR_SUCCESS") both return ERROR_SUCCESS.
func FromName(name string) (WIN32_ERROR, bool) {
	code, defined := nameToCode[name]
	return code, defined
}

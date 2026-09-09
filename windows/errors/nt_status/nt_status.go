// Package nt_status is the NTSTATUS table of [MS-ERREF] section 2.3.1, extended
// with the CIFS-specific values [MS-CIFS] adds to it.
//
// An NTSTATUS is a bitfield: severity in bits 31:30, a customer flag, a reserved
// bit, a 12-bit facility and a 16-bit code. Severity is what makes
// NT_STATUS_BUFFER_OVERFLOW (0x80000005) a warning that carries data rather than
// a failure, and it is why these values cannot be read as Win32 error codes,
// which are flat integers with no such structure. The two spaces also collide by
// value: 0x00000002 is NT_STATUS_WAIT_2 here and ERROR_FILE_NOT_FOUND in
// [github.com/TheManticoreProject/Manticore/windows/errors/win32], and
// 0x00000103 is NT_STATUS_PENDING here and ERROR_NO_MORE_ITEMS there. Which
// table applies is a property of the field being read, so the two are separate
// types.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-erref/596a1078-e883-4972-9bbc-49e60bebca55
package nt_status

import "fmt"

//go:generate go run ../errgen -in ../errgen/ms-erref-2.3.1-ntstatus.tsv -out . -package nt_status -type NT_STATUS -strip-name-prefix STATUS_ -add-name-prefix NT_STATUS_ -source https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-erref/596a1078-e883-4972-9bbc-49e60bebca55

// NT_STATUS is an NTSTATUS value.
type NT_STATUS uint32

// Entry is what the specification records for one status.
type Entry struct {
	// Name is the status's canonical Go constant name, such as
	// "NT_STATUS_ACCESS_DENIED".
	Name string
	// Description is the description the specification gives, verbatim. A
	// percentage sign followed by alphanumeric characters — "%1", "%hs" — marks
	// a variable that Windows replaces with text when the status is returned.
	Description string
}

// lookup resolves a status against the CIFS extensions first and the [MS-ERREF]
// table second.
//
// The order matters for exactly one value. 0x00010002 is NT_STATUS_INVALID_SMB
// in [MS-CIFS] and NT_STATUS_DBG_CONTINUE in [MS-ERREF], because the CIFS values
// are ErrorClass | ErrorCode<<16 composites that land in the same numeric region
// as the [MS-ERREF] DBG_* block. This tree produces the SMB one and never the
// debugger one, so the SMB name is the useful answer; TestCIFSStatusPrecedence
// pins that this is the only value where the choice arises.
func lookup(s NT_STATUS) (Entry, bool) {
	if entry, defined := cifsTable[s]; defined {
		return entry, true
	}
	entry, defined := table[s]
	return entry, defined
}

// String returns the status's constant name, or its hexadecimal value when
// neither specification defines a name for it. It never returns the empty
// string, so it is safe to interpolate into a log line directly.
func (s NT_STATUS) String() string {
	if entry, defined := lookup(s); defined {
		return entry.Name
	}
	return fmt.Sprintf("0x%08x", uint32(s))
}

// Name returns the status's canonical constant name, or the empty string when
// neither specification defines a name for it. Use [NT_STATUS.String] to render
// a status whose name may be unknown.
func (s NT_STATUS) Name() string {
	entry, _ := lookup(s)
	return entry.Name
}

// Description returns the description the specification gives for the status, or
// the empty string when it defines none.
func (s NT_STATUS) Description() string {
	entry, _ := lookup(s)
	return entry.Description
}

// IsSuccess reports whether the status is NT_STATUS_SUCCESS, which is the
// condition under which [NT_STATUS.Error] reports no error.
//
// It is narrower than the severity field: NTSTATUS defines a whole success
// severity and an informational one, so values such as NT_STATUS_PENDING and
// NT_STATUS_BUFFER_OVERFLOW are not failures in the specification's sense but
// are not NT_STATUS_SUCCESS either, and a caller that treats them as success
// must say so itself.
func (s NT_STATUS) IsSuccess() bool {
	return s == NT_STATUS_SUCCESS
}

// Error returns nil for NT_STATUS_SUCCESS and an error for every other status,
// including one neither specification defines — an undefined non-success status
// is still a failure, and reporting nil for it would read as success.
func (s NT_STATUS) Error() error {
	if s.IsSuccess() {
		return nil
	}
	if entry, defined := lookup(s); defined {
		return fmt.Errorf("NT_STATUS(0x%08x): %s: %s", uint32(s), entry.Name, entry.Description)
	}
	return fmt.Errorf("NT_STATUS(0x%08x)", uint32(s))
}

// Lookup returns the specification's entry for a status, and whether either
// specification defines one.
func Lookup(status NT_STATUS) (Entry, bool) {
	return lookup(status)
}

// FromName resolves a symbolic name to its status, and reports whether the name
// is defined. Both the Go constant name and the name the specification uses
// resolve, so FromName("NT_STATUS_ACCESS_DENIED") and
// FromName("STATUS_ACCESS_DENIED") both return NT_STATUS_ACCESS_DENIED.
func FromName(name string) (NT_STATUS, bool) {
	if status, defined := cifsNameToCode[name]; defined {
		return status, true
	}
	status, defined := nameToCode[name]
	return status, defined
}

package hresult

// The two most common HRESULT values are absent from the [MS-ERREF] 2.1.1 table
// that generates codes.go: it starts at 0x00030200 and lists no value below
// that, so neither S_OK nor S_FALSE appears in it. They are declared here by
// hand, the way the [MS-CIFS] extensions to the NTSTATUS table are.
//
// Both are success-severity values, so [HRESULT.Error] reports nil for each.
// S_FALSE is the one that catches callers out: it means the call succeeded and
// the answer was "no", not that the call failed.
//
// Source: [MS-ERREF] 2.1, and the S_OK/S_FALSE definitions in winerror.h.
const (
	// S_OK is success. The HRESULT_FROM_WIN32 macro maps ERROR_SUCCESS to it.
	S_OK HRESULT = 0x00000000
	// S_FALSE is success with a negative answer — the call completed and
	// returned "no".
	S_FALSE HRESULT = 0x00000001
)

// wellKnown holds the entries for the values [MS-ERREF] 2.1.1 omits, consulted
// by lookup after the generated table.
var wellKnown = map[HRESULT]Entry{
	S_OK:    {Name: "S_OK", Description: "Operation successful."},
	S_FALSE: {Name: "S_FALSE", Description: "Operation successful but returned no results."},
}

// wellKnownNames resolves those names for FromName.
var wellKnownNames = func() map[string]HRESULT {
	names := make(map[string]HRESULT, len(wellKnown))
	for value, entry := range wellKnown {
		names[entry.Name] = value
	}
	return names
}()

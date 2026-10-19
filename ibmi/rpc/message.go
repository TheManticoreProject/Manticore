package rpc

import "strconv"

// IBMMessage is one IBM i message returned by a call, independent of the
// transport that carried it.
//
// Text and SubstitutionData can contain sensitive values such as object or
// profile names; String omits them, and diagnostics should log only ID and
// Severity.
type IBMMessage struct {
	// ID is the seven-character message identifier, for example "CPF9801".
	ID string
	// Type is the IBM message type code.
	Type uint16
	// Severity is the IBM severity, 0 to 99.
	Severity uint16
	// FileName and LibraryName identify the message file.
	FileName    string
	LibraryName string
	// Text is the first-level message text, decoded from the job CCSID.
	Text string
	// Help is the second-level (help) text, when the server's message
	// format carries it (Remote Command level 10 and later); else "".
	Help string
	// SubstitutionData is the raw replacement data for the message text.
	SubstitutionData []byte
}

// String returns only the safe identity of the message, for example
// "CPF9801 (severity 40)".
func (m IBMMessage) String() string {
	return m.ID + " (severity " + strconv.Itoa(int(m.Severity)) + ")"
}

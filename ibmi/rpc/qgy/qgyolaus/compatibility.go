package qgyolaus

// Compatibility records the evidence behind each supported format. Equal
// record lengths across releases are not proof of identical meaning, so a
// format counts as verified for V5R4 only with V5R4-era evidence.
type Compatibility struct {
	Format       string
	RecordLength uint32
	// V5R4 is true when V5R4-era evidence establishes the layout.
	V5R4     bool
	Evidence []string
}

// SupportedFormats lists the formats this package decodes, with evidence.
func SupportedFormats() []Compatibility {
	ibm76 := "IBM i 7.6 QGYOLAUS documentation (retrieved 2026-10-14)"
	return []Compatibility{
		{"AUTU0100", 12, true, []string{ibm76,
			"installed V5R4 Navigator PCML HostProfileQgyolaus: 12-byte AUTU0100 record (iSeriesNavigatorReverse docs/37, .research/family-af)"}},
		{"AUTU0150", 62, true, []string{ibm76,
			"live V5R4 run of the proof of concept returned 63 AUTU0150 records of 62 bytes (iSeriesNavigatorReverse docs/57)"}},
		{"AUTU0200", 176, true, []string{ibm76,
			"IBM V5R3 and V5R4 Security-related APIs manuals define the same 176-byte layout (iSeriesNavigatorReverse docs/58)"}},
		{"AUTU0250", 228, false, []string{ibm76}},
	}
}

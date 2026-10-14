package rmtcmd

// Capabilities are the Remote Command features implied by the negotiated
// data-stream level. Codecs consult these flags instead of comparing levels.
type Capabilities struct {
	RemoteCommandLevel uint16
	// InputOutputUsage33 selects program-call usage 33 for input/output
	// parameters (level 5 and later); below it usage 13 is used.
	InputOutputUsage33 bool
	// AllCommandMessages means the "all messages" option is honored (level 7
	// and later); below it requests are downgraded to "up to ten".
	AllCommandMessages bool
	// ExplicitCommandCCSID means Run Command carries an explicit command
	// CCSID in CP 0x1104 (level 10 and later).
	ExplicitCommandCCSID bool
}

// CapabilitiesForLevel derives the capabilities of a data-stream level.
func CapabilitiesForLevel(level uint16) Capabilities {
	return Capabilities{
		RemoteCommandLevel:   level,
		InputOutputUsage33:   level >= 5,
		AllCommandMessages:   level >= 7,
		ExplicitCommandCCSID: level >= 10,
	}
}

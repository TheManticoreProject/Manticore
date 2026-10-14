package hostserver

// NewCorrelatorAt returns a Correlator whose last issued ID is last, so tests
// can reach the exhaustion boundary without 4 billion calls.
func NewCorrelatorAt(last uint32) *Correlator { return &Correlator{last: last} }

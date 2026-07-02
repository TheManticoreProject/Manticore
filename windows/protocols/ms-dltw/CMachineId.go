package msdltw

import "bytes"

// CMachineId is a MachineID ([MS-DLTW] 2.2.4): the identifier of a computer that
// participates in link tracking. It is a fixed 16-octet field (the IDL member
// `char _szMachine[16]`) holding a null-terminated ASCII string that names the computer;
// unused trailing octets are zero. The array is transmitted inline under NDR (a fixed
// array carries no conformance/variance header).
type CMachineId struct {
	Machine [16]byte
}

// String returns the machine name as a Go string, trimming at the first NUL terminator.
func (m CMachineId) String() string {
	if i := bytes.IndexByte(m.Machine[:], 0); i >= 0 {
		return string(m.Machine[:i])
	}
	return string(m.Machine[:])
}

// NewCMachineId builds a CMachineId from a name, copying at most 15 octets so the fixed
// 16-octet field always retains a NUL terminator (matching the null-terminated ASCII
// string the wire format carries).
func NewCMachineId(name string) CMachineId {
	var m CMachineId
	copy(m.Machine[:15], name)
	return m
}

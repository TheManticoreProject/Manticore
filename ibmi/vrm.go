package ibmi

import (
	"errors"
	"fmt"
)

// ErrInvalidVRM is matched (via errors.Is) by every VRM construction or
// parsing failure.
var ErrInvalidVRM = errors.New("ibmi: invalid VRM")

// VRM is a validated IBM i release identity: version, release and
// modification, each a single decimal digit, as in V5R4M0. Version 0 does
// not exist, so the zero VRM is never valid and reports IsZero.
//
// VRM is a comparable value type; use NewVRM or ParseVRM to build one.
type VRM struct {
	value uint16 // version*100 + release*10 + modification
}

// NewVRM returns the VRM for the given components. Version must be 1..9;
// release and modification must be 0..9.
func NewVRM(version, release, modification uint8) (VRM, error) {
	if version < 1 || version > 9 || release > 9 || modification > 9 {
		return VRM{}, fmt.Errorf("%w: V%dR%dM%d (version must be 1-9, release and modification 0-9)",
			ErrInvalidVRM, version, release, modification)
	}
	return VRM{value: uint16(version)*100 + uint16(release)*10 + uint16(modification)}, nil
}

// ParseVRM parses the canonical form "VxRyMz", for example "V5R4M0". The
// letters must be uppercase and each component exactly one digit.
func ParseVRM(text string) (VRM, error) {
	if len(text) != 6 || text[0] != 'V' || text[2] != 'R' || text[4] != 'M' ||
		!isDigit(text[1]) || !isDigit(text[3]) || !isDigit(text[5]) {
		return VRM{}, fmt.Errorf("%w: %q is not of the form VxRyMz", ErrInvalidVRM, text)
	}
	return NewVRM(text[1]-'0', text[3]-'0', text[5]-'0')
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }

// Version returns the version component (1..9), or 0 for the zero VRM.
func (v VRM) Version() uint8 { return uint8(v.value / 100) }

// Release returns the release component (0..9).
func (v VRM) Release() uint8 { return uint8(v.value / 10 % 10) }

// Modification returns the modification component (0..9).
func (v VRM) Modification() uint8 { return uint8(v.value % 10) }

// IsZero reports whether v is the zero VRM, which no constructor returns.
func (v VRM) IsZero() bool { return v.value == 0 }

// Compare returns -1, 0 or +1 as v is older than, equal to or newer than
// other.
func (v VRM) Compare(other VRM) int {
	switch {
	case v.value < other.value:
		return -1
	case v.value > other.value:
		return 1
	}
	return 0
}

// AtLeast reports whether v is the same release as other or newer.
func (v VRM) AtLeast(other VRM) bool { return v.value >= other.value }

// String returns the canonical "VxRyMz" form, or "V?R?M?" for the zero VRM.
func (v VRM) String() string {
	if v.IsZero() {
		return "V?R?M?"
	}
	return fmt.Sprintf("V%dR%dM%d", v.Version(), v.Release(), v.Modification())
}

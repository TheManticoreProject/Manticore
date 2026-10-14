package rpc

import "fmt"

// MaxObjectNameLength is the length of an IBM i system object name.
const MaxObjectNameLength = 10

// ObjectName is a validated IBM i library or program name.
//
// The first milestone accepts only the unquoted system-name subset that is a
// single byte in both supported job CCSIDs (37 and 1140): 1 to 10 characters,
// the first from A-Z, $, # or @, the rest from A-Z, 0-9, $, #, @ or _. ASCII
// lowercase letters are uppercased. Anything else, including quoted names,
// special values such as *LIBL, and national characters, is rejected rather
// than truncated or repaired.
type ObjectName struct {
	name string
}

// NewObjectName validates and canonicalizes name.
func NewObjectName(name string) (ObjectName, error) {
	if len(name) == 0 || len(name) > MaxObjectNameLength {
		return ObjectName{}, invalid("object name", fmt.Sprintf("%q must be 1 to %d characters", name, MaxObjectNameLength))
	}
	out := make([]byte, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		ok := (c >= 'A' && c <= 'Z') || c == '$' || c == '#' || c == '@'
		if i > 0 {
			ok = ok || (c >= '0' && c <= '9') || c == '_'
		}
		if !ok {
			return ObjectName{}, invalid("object name", fmt.Sprintf("%q has a character outside the supported system-name set at position %d", name, i))
		}
		out[i] = c
	}
	return ObjectName{name: string(out)}, nil
}

// String returns the canonical uppercase name.
func (n ObjectName) String() string { return n.name }

// IsZero reports whether n is the zero value, which no constructor returns.
func (n ObjectName) IsZero() bool { return n.name == "" }

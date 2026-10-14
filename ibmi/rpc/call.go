package rpc

import "fmt"

// MessageOption selects how many IBM i messages a call returns.
type MessageOption uint8

const (
	// MessagesUpToTen returns at most ten messages.
	MessagesUpToTen MessageOption = iota + 1
	// MessagesNone returns no messages.
	MessagesNone
	// MessagesAll returns every message. A transport may downgrade it to
	// MessagesUpToTen and report that in Result.EffectiveMessageOption.
	MessagesAll
)

func (o MessageOption) String() string {
	switch o {
	case MessagesUpToTen:
		return "up-to-ten"
	case MessagesNone:
		return "none"
	case MessagesAll:
		return "all"
	}
	return fmt.Sprintf("MessageOption(%d)", uint8(o))
}

// Valid reports whether o is one of the defined options.
func (o MessageOption) Valid() bool { return o >= MessagesUpToTen && o <= MessagesAll }

// Call is an immutable, validated IBM i program call: a qualified program
// and its ordered parameters. It describes the call semantically; a Caller
// owns the wire encoding. The zero Call is invalid.
type Call struct {
	library    ObjectName
	program    ObjectName
	parameters []Parameter
	option     MessageOption
}

// NewCall validates and returns a call to library/program. The parameter
// slice is copied. Zero parameters are allowed; a zero Parameter is not.
func NewCall(library, program ObjectName, parameters []Parameter, option MessageOption) (Call, error) {
	if library.IsZero() {
		return Call{}, invalid("library", "object name is required")
	}
	if program.IsZero() {
		return Call{}, invalid("program", "object name is required")
	}
	if !option.Valid() {
		return Call{}, invalid("message option", option.String()+" is not defined")
	}
	for i, p := range parameters {
		if p.IsZero() {
			return Call{}, invalid(fmt.Sprintf("parameter[%d]", i), "zero Parameter; use a constructor")
		}
	}
	return Call{
		library:    library,
		program:    program,
		parameters: append([]Parameter(nil), parameters...),
		option:     option,
	}, nil
}

// Library returns the program's library.
func (c Call) Library() ObjectName { return c.library }

// Program returns the program name.
func (c Call) Program() ObjectName { return c.program }

// Parameters returns a copy of the ordered parameters.
func (c Call) Parameters() []Parameter { return append([]Parameter(nil), c.parameters...) }

// ParameterCount returns the number of parameters.
func (c Call) ParameterCount() int { return len(c.parameters) }

// MessageOption returns the requested message option.
func (c Call) MessageOption() MessageOption { return c.option }

// IsZero reports whether c is the zero Call, which NewCall never returns.
func (c Call) IsZero() bool { return c.program.IsZero() }

// String describes the call without revealing parameter data, for example
// "QGY/QGYOLAUS(7 parameters)".
func (c Call) String() string {
	return fmt.Sprintf("%s/%s(%d parameters)", c.library, c.program, len(c.parameters))
}

// GoString keeps %#v from printing parameter data through the unexported
// parameter slice.
func (c Call) GoString() string { return "rpc.Call{" + c.String() + "}" }

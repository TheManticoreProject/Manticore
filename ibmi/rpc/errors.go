package rpc

import "errors"

// ErrInvalid is matched (via errors.Is) by every validation failure in this
// package: a bad object name, parameter, call or result. Validation happens
// in constructors, so an invalid call never reaches a transport.
var ErrInvalid = errors.New("rpc: validation failed")

// ValidationError identifies which value failed validation and why. It never
// carries parameter data.
type ValidationError struct {
	Field  string // for example "program", "parameter[2]", "result.parameters"
	Reason string
}

func (e *ValidationError) Error() string {
	return "rpc: invalid " + e.Field + ": " + e.Reason
}

// Is makes errors.Is(err, ErrInvalid) true.
func (e *ValidationError) Is(target error) bool { return target == ErrInvalid }

func invalid(field, reason string) error {
	return &ValidationError{Field: field, Reason: reason}
}

package rpc

import "fmt"

// ParameterResult is what a call returned for one request parameter.
// Present distinguishes a slot that returned nothing (input-only parameters,
// or outputs the server did not return) from an output that came back with
// zero bytes.
type ParameterResult struct {
	Present bool
	Data    []byte
}

// Result is the transport-neutral outcome of a Call. Parameters aligns
// one-for-one with the call's parameters. A Caller may return a Result
// together with a non-nil error when IBM i completed the call with a failure
// status; Messages then explains the failure.
type Result struct {
	Parameters             []ParameterResult
	Messages               []IBMMessage
	RequestedMessageOption MessageOption
	EffectiveMessageOption MessageOption
	// CorrelationID identifies the transport exchange, or 0 if the
	// transport has none.
	CorrelationID uint32
}

// Validate checks that r is a well-formed result of call: one slot per
// parameter, no data for input-only slots, no output beyond the requested
// maximum, the call's message option echoed as requested, and a defined
// effective option. Callers and their fakes should run it before returning.
func (r *Result) Validate(call Call) error {
	if r == nil {
		return invalid("result", "nil")
	}
	if len(r.Parameters) != len(call.parameters) {
		return invalid("result.parameters", fmt.Sprintf("%d slots for %d parameters", len(r.Parameters), len(call.parameters)))
	}
	for i, slot := range r.Parameters {
		p := call.parameters[i]
		field := fmt.Sprintf("result.parameters[%d]", i)
		if !slot.Present {
			if len(slot.Data) != 0 {
				return invalid(field, "data without Present")
			}
			continue
		}
		if !p.direction.returnsData() {
			return invalid(field, "input-only parameter returned data")
		}
		if uint64(len(slot.Data)) > uint64(p.maximum) {
			return invalid(field, fmt.Sprintf("%d bytes exceed the %d-byte maximum", len(slot.Data), p.maximum))
		}
	}
	if r.RequestedMessageOption != call.option {
		return invalid("result.requested message option", fmt.Sprintf("%s, call requested %s", r.RequestedMessageOption, call.option))
	}
	if !r.EffectiveMessageOption.Valid() {
		return invalid("result.effective message option", r.EffectiveMessageOption.String()+" is not defined")
	}
	return nil
}

// Output returns the data of parameter i and whether it is present. It
// returns (nil, false) for an out-of-range index.
func (r *Result) Output(i int) ([]byte, bool) {
	if r == nil || i < 0 || i >= len(r.Parameters) {
		return nil, false
	}
	slot := r.Parameters[i]
	return slot.Data, slot.Present
}

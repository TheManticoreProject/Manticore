package rpc

import (
	"fmt"
	"math"
)

// Direction is how a program parameter is passed.
type Direction uint8

const (
	// DirectionInput sends data and expects nothing back.
	DirectionInput Direction = iota + 1
	// DirectionOutput allocates a receiver of a maximum size.
	DirectionOutput
	// DirectionInputOutput sends data in a receiver of a maximum size and
	// expects it back.
	DirectionInputOutput
)

func (d Direction) String() string {
	switch d {
	case DirectionInput:
		return "input"
	case DirectionOutput:
		return "output"
	case DirectionInputOutput:
		return "input/output"
	}
	return fmt.Sprintf("Direction(%d)", uint8(d))
}

// returnsData reports whether the direction produces output.
func (d Direction) returnsData() bool {
	return d == DirectionOutput || d == DirectionInputOutput
}

// Parameter is one immutable program parameter. Constructors copy their
// input, and no method exposes the stored bytes, so a Parameter cannot be
// changed after it is built. The zero Parameter is invalid.
type Parameter struct {
	direction Direction
	data      []byte
	maximum   uint32
}

// checkInputLength rejects input larger than the 32-bit parameter length.
func checkInputLength(n uint64) error {
	if n > math.MaxUint32 {
		return invalid("parameter", fmt.Sprintf("input of %d bytes exceeds the %d-byte maximum", n, uint64(math.MaxUint32)))
	}
	return nil
}

// Input returns an input-only parameter holding a copy of data. A nil or
// empty data is an explicit zero-length input, as with EmptyInput.
func Input(data []byte) (Parameter, error) {
	if err := checkInputLength(uint64(len(data))); err != nil {
		return Parameter{}, err
	}
	return Parameter{direction: DirectionInput, data: clone(data), maximum: uint32(len(data))}, nil
}

// EmptyInput returns an explicit zero-length input parameter.
func EmptyInput() Parameter {
	return Parameter{direction: DirectionInput, data: []byte{}}
}

// Output returns an output-only parameter with a receiver of maximum bytes.
// A zero maximum is rejected: it would be ambiguous with "no receiver".
func Output(maximum uint32) (Parameter, error) {
	if maximum == 0 {
		return Parameter{}, invalid("parameter", "output maximum must be nonzero")
	}
	return Parameter{direction: DirectionOutput, maximum: maximum}, nil
}

// InputOutput returns a parameter that sends a copy of data in a receiver of
// maximum bytes. maximum must be nonzero and at least len(data).
func InputOutput(data []byte, maximum uint32) (Parameter, error) {
	if err := checkInputLength(uint64(len(data))); err != nil {
		return Parameter{}, err
	}
	if maximum == 0 {
		return Parameter{}, invalid("parameter", "input/output maximum must be nonzero")
	}
	if uint64(len(data)) > uint64(maximum) {
		return Parameter{}, invalid("parameter", fmt.Sprintf("input of %d bytes exceeds the %d-byte input/output maximum", len(data), maximum))
	}
	return Parameter{direction: DirectionInputOutput, data: clone(data), maximum: maximum}, nil
}

// Direction returns how the parameter is passed.
func (p Parameter) Direction() Direction { return p.direction }

// Maximum returns the receiver size for output-capable parameters, or the
// input length for input-only parameters.
func (p Parameter) Maximum() uint32 { return p.maximum }

// InputLength returns the number of input bytes (zero for output-only).
func (p Parameter) InputLength() uint32 { return uint32(len(p.data)) }

// InputBytes returns a copy of the input data: non-nil for input and
// input/output parameters, nil for output-only parameters.
func (p Parameter) InputBytes() []byte {
	if p.direction == DirectionOutput {
		return nil
	}
	return clone(p.data)
}

// IsZero reports whether p is the zero Parameter, which no constructor
// returns.
func (p Parameter) IsZero() bool { return p.direction == 0 }

// String describes the parameter without revealing its data.
func (p Parameter) String() string {
	switch p.direction {
	case DirectionInput:
		return fmt.Sprintf("input(%d bytes)", len(p.data))
	case DirectionOutput:
		return fmt.Sprintf("output(max %d)", p.maximum)
	case DirectionInputOutput:
		return fmt.Sprintf("input/output(%d bytes, max %d)", len(p.data), p.maximum)
	}
	return "invalid"
}

// GoString keeps %#v from printing the parameter data.
func (p Parameter) GoString() string { return "rpc.Parameter{" + p.String() + "}" }

func clone(b []byte) []byte {
	out := make([]byte, len(b))
	copy(out, b)
	return out
}

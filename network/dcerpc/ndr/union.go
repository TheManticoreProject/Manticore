package ndr

import "reflect"

// Union is implemented by an NDR encapsulated union: a discriminant (the switch
// value) followed by the arm it selects. The walker writes the discriminant inline
// (a 4-octet switch, the common MS-RPCE case) and then delegates the arm to the type.
//
// References:
//   - [C706] section 14.3.8 (Unions):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap14.htm
//   - [MS-RPCE] 2.2.4.x Union types:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/b6090c2b-f44a-47a1-a13b-b82ade0137b2
//
// Non-encapsulated unions (discriminant carried by a separate switch_is parameter,
// not inline) are not yet handled by the walker.
type Union interface {
	// SwitchValue returns the discriminant for the currently-set arm.
	SwitchValue() uint32
	// MarshalArm marshals the arm selected by sw.
	MarshalArm(e *Encoder, sw uint32) error
	// UnmarshalArm reads the arm selected by sw.
	UnmarshalArm(d *Decoder, sw uint32) error
}

var unionType = reflect.TypeOf((*Union)(nil)).Elem()

// asUnion returns fv (or its address) as a Union, if it implements one.
func asUnion(fv reflect.Value) (Union, bool) {
	if fv.Type().Implements(unionType) {
		return fv.Interface().(Union), true
	}
	if fv.CanAddr() && fv.Addr().Type().Implements(unionType) {
		return fv.Addr().Interface().(Union), true
	}
	return nil, false
}

package ndr

import "reflect"

// NDR discriminated unions, declared with struct tags rather than the Marshaler escape
// hatch. A union is a Go struct with one field tagged `switch` (the discriminant) and
// one field per arm tagged `case=<value>` (an optional `default` arm covers the rest):
//
//	type LSAPR_POLICY_INFORMATION struct {
//	    Level     uint32                   `ndr:"switch"`
//	    AuditLog  *PolicyAuditLogInfo      `ndr:"case=1,unique"`
//	    AuditFull *PolicyAuditFullInfo     `ndr:"case=2,unique"`
//	    Default   *PolicyDefaultInfo       `ndr:"default,unique"`
//	}
//
// The wire form is the discriminant followed by the single selected arm ([C706] section
// 14.3.8: "NDR represents a union as a representation of the tag followed by a
// representation of the selected member"). The discriminant is always transmitted as
// the first part of the union representation, even for a non-encapsulated union whose
// IDL discriminant is a separate `switch_is` parameter — for such a union [C706] §14.3.8
// transmits the value twice, once as the external field and once inline here, so a
// receiver decodes the arm from this inline tag without needing the external parameter
// (which, for an [out] union, is not even present in the response). The union is aligned
// to the largest alignment of the discriminant and any arm, matching the static union
// alignment used by Windows/[MS-RPCE] implementations (a receiver cannot know the
// selected arm before reading the tag, so the padding must not depend on it).
//
// References:
//   - [C706] section 14.3.8 (Unions):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap14.htm
//   - [MS-RPCE] 2.2.4 NDR Transfer Syntax:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/b6090c2b-f44a-47a1-a13b-b82ade0137b2
//   - union attribute ([MS] MIDL non-encapsulated unions):
//     https://learn.microsoft.com/en-us/windows/win32/midl/union

// unionSwitchIndex returns the index of the field tagged `switch` (the union
// discriminant), or -1 if the struct is not a declarative union.
func unionSwitchIndex(rt reflect.Type) int {
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.PkgPath != "" {
			continue
		}
		if parseTag(f.Tag.Get("ndr")).isSwitch {
			return i
		}
	}
	return -1
}

// unionArm returns the index of the arm selected by the discriminant value, falling
// back to the `default` arm, or -1 if neither matches (an arm-less discriminant value,
// which is valid: the union body is then empty).
func unionArm(rt reflect.Type, disc reflect.Value) int {
	var dv int64
	if isUintKind(disc.Kind()) {
		dv = int64(disc.Uint())
	} else {
		dv = disc.Int()
	}
	def := -1
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tag := parseTag(f.Tag.Get("ndr"))
		if tag.isDefault {
			def = i
		}
		if tag.hasCase && tag.caseVal == dv {
			return i
		}
	}
	return def
}

// unionArmAlignment returns the largest NDR alignment among the union's arms (the
// case/default fields), which is the alignment applied to the selected arm so its
// position does not depend on which arm was sent ([C706] section 14.3.8).
func unionArmAlignment(rt reflect.Type) int {
	a := 1
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tag := parseTag(f.Tag.Get("ndr"))
		if !tag.hasCase && !tag.isDefault {
			continue
		}
		if fa := ndrAlignment(f.Type); fa > a {
			a = fa
		}
	}
	return a
}

// marshalUnion writes the discriminant then the single arm it selects. The arm is
// written through the inline-field path so a pointer arm (the common case — each arm is
// a [unique]/[ref] pointer to a structure) emits a referent id with its body deferred to
// the end of the union construction.
func marshalUnion(e *Encoder, rv reflect.Value, swIdx int) error {
	rt := rv.Type()
	disc := rv.Field(swIdx)
	// The discriminant is aligned to its own alignment. The selected arm is then aligned
	// to the LARGEST alignment among all arms — not just the selected one — so a receiver
	// can position the arm before it knows which arm was sent. Verified on the wire:
	// LSAPR_POLICY_INFORMATION's 2-byte server-role arm still starts 8-aligned because
	// other arms carry 8-byte LARGE_INTEGER members. (Note this padding follows the tag,
	// so it does not over-align the discriminant itself.)
	e.Align(ndrAlignment(disc.Type()))
	if err := marshalScalar(e, disc); err != nil {
		return err
	}
	armIdx := unionArm(rt, disc)
	if armIdx < 0 {
		return nil // discriminant value with no arm and no default: empty body
	}
	e.Align(unionArmAlignment(rt))
	f := rt.Field(armIdx)
	var deferred []func() error
	if err := marshalFieldInline(e, rv.Field(armIdx), parseTag(f.Tag.Get("ndr")), &deferred, true); err != nil {
		return err
	}
	return runDeferred(deferred)
}

// unmarshalUnion reads the discriminant, selects the matching arm, and leaves the other
// arm fields zero.
func unmarshalUnion(d *Decoder, rv reflect.Value, swIdx int) error {
	rt := rv.Type()
	disc := rv.Field(swIdx)
	d.Align(ndrAlignment(disc.Type())) // discriminant aligned to itself
	if err := unmarshalScalar(d, disc); err != nil {
		return err
	}
	armIdx := unionArm(rt, disc)
	if armIdx < 0 {
		return nil
	}
	d.Align(unionArmAlignment(rt)) // arm aligned to the largest alignment among all arms
	f := rt.Field(armIdx)
	var deferred []func() error
	if err := unmarshalFieldInline(d, rv.Field(armIdx), parseTag(f.Tag.Get("ndr")), &deferred, true); err != nil {
		return err
	}
	return runDeferred(deferred)
}

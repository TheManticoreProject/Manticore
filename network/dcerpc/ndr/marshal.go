package ndr

import (
	"fmt"
	"reflect"
)

// Marshal encodes v (a struct or pointer to struct) into an NDR octet stream. Each
// exported field is marshalled in declaration order according to its `ndr:"..."` tag.
//
// Pointers (Go pointer fields, or value fields tagged "unique"/"ptr") are encoded as a
// referent id inline with their referent body deferred to the end of the enclosing
// construction, per the deferral rule in [C706] section 14.3.10:
// https://pubs.opengroup.org/onlinepubs/9629399/chap14.htm
func Marshal(v any) ([]byte, error) { return MarshalAs(v, NDR20) }

// MarshalAs is Marshal under an explicit transfer syntax. NDR64 widens counts and
// referent ids to 8 octets and aligns them to 8 ([MS-RPCE] section 2.2.5). NDR64
// unions and pipes are not yet supported and marshalling a value that contains one
// returns an error rather than emitting an unverified encoding.
func MarshalAs(v any, syntax Syntax) ([]byte, error) {
	rv, err := structValue(reflect.ValueOf(v))
	if err != nil {
		return nil, err
	}
	e := NewEncoderForSyntax(syntax)
	if err := marshalStruct(e, rv, false); err != nil {
		return nil, err
	}
	return e.Bytes(), nil
}

// Unmarshal decodes an NDR octet stream into v, which must be a non-nil pointer to a
// struct.
func Unmarshal(data []byte, v any) error { return UnmarshalAs(data, v, NDR20) }

// UnmarshalAs is Unmarshal under an explicit transfer syntax. See MarshalAs for the
// NDR64 semantics and the union/pipe limitation.
func UnmarshalAs(data []byte, v any, syntax Syntax) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("ndr: Unmarshal requires a non-nil pointer, got %T", v)
	}
	sv, err := structValue(rv)
	if err != nil {
		return err
	}
	return unmarshalStruct(NewDecoderForSyntax(data, syntax), sv, false)
}

// structValue dereferences pointers/interfaces to reach a struct value.
func structValue(rv reflect.Value) (reflect.Value, error) {
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return reflect.Value{}, fmt.Errorf("ndr: nil value")
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("ndr: expected a struct, got %s", rv.Kind())
	}
	return rv, nil
}

// ---- marshalling ----------------------------------------------------------------

// marshalStruct marshals a struct as an NDR construction: all inline field
// representations first, then the deferred referent bodies in field order.
//
// If the struct embeds a conformant array directly (a non-pointer byte slice), its
// maximum_count is hoisted to the start of the struct and only the elements are
// emitted in place, per [MS-RPCE] "Structure Containing a Conformant Varying Array":
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/7bae54af-8b6a-4672-a6a3-6bcdf131730a
// embedded is true when rv is not the top-level call/parameter struct, which governs
// how [ref] pointer members are encoded ([C706] section 14.3.10).
func marshalStruct(e *Encoder, rv reflect.Value, embedded bool) error {
	var deferred []func() error
	retvalIdx, err := marshalStructFields(e, rv, embedded, &deferred)
	if err != nil {
		return err
	}
	if err := runDeferred(deferred); err != nil {
		return err
	}
	// The RPC return value follows all [out] parameters and their deferred referents.
	if retvalIdx >= 0 {
		rt := rv.Type()
		f := rt.Field(retvalIdx)
		var rdef []func() error
		if err := marshalFieldInline(e, rv.Field(retvalIdx), parseTag(f.Tag.Get("ndr")), &rdef, embedded); err != nil {
			return fmt.Errorf("ndr: field %s: %w", f.Name, err)
		}
		if err := runDeferred(rdef); err != nil {
			return err
		}
	}
	return nil
}

// marshalStructFields writes a struct's inline representation, appending each deferred
// referent body to *deferred rather than flushing it. The caller owns the flush, which
// is what lets an enclosing array emit every element's referents after the whole array
// body, per [C706] section 14.3.10. It returns the index of the `retval` field (or -1),
// which the caller emits last; array elements never carry a retval, so callers other
// than marshalStruct ignore it.
func marshalStructFields(e *Encoder, rv reflect.Value, embedded bool, deferred *[]func() error) (int, error) {
	rt := rv.Type()
	// A structure is aligned to the largest alignment of any of its members before its
	// representation ([C706] section 14.2.2). The first member self-aligns to its own
	// size, which is insufficient when a later member is wider (e.g. RPC_UNICODE_STRING,
	// whose Buffer pointer makes it 4-aligned, following a 2-byte discriminant).
	confIdx := embeddedConformantIndex(rt, rv)
	e.Align(leadingAlignment(rt, confIdx, e.syntax))
	if confIdx >= 0 {
		// The hoisted maximum_count is a size determinant aligned to the determinant
		// width (4 under NDR20), NOT the element alignment. After it, the structure body
		// (its members) begins at the structure's natural alignment ([C706] 14.2.2) —
		// which, for a conformant array of 8-octet-aligned elements (e.g.
		// PROPERTY_META_DATA_EXT), is 8. So re-align before the members; without this the
		// member that follows is written 4-aligned and every record drifts.
		e.writeCount(uint64(rv.Field(confIdx).Len())) // hoisted maximum_count
		e.Align(ndrAlignment(rv.Field(confIdx).Type().Elem(), e.syntax))
	}
	// A field named by another field's size_is/length_is is the array's count; its
	// value is derived from the array length so the count and the elements cannot
	// disagree and the caller need not set it explicitly.
	counts := countTargets(rt, rv)
	retvalIdx := retvalIndex(rt)
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.PkgPath != "" { // unexported
			continue
		}
		tag := parseTag(f.Tag.Get("ndr"))
		if tag.skip {
			continue
		}
		resolveSiblingCounts(&tag, rt, rv)
		if i == retvalIdx {
			continue // the RPC return value is encoded last, after deferred referents
		}
		if c, ok := counts[f.Name]; ok && isUintKind(rv.Field(i).Kind()) {
			if tag.align > 0 {
				e.Align(tag.align)
			}
			tmp := reflect.New(f.Type).Elem()
			tmp.SetUint(c)
			if err := marshalScalar(e, tmp); err != nil {
				return -1, fmt.Errorf("ndr: field %s: %w", f.Name, err)
			}
			continue
		}
		if i == confIdx {
			// The maximum_count was hoisted above. For a conformant-varying array the
			// offset and actual_count stay here, ahead of the elements.
			if tag.varying {
				e.writeCount(0)                         // offset
				e.writeCount(uint64(rv.Field(i).Len())) // actual_count
			}
			if err := marshalElements(e, rv.Field(i), elementTag(tag)); err != nil {
				return -1, fmt.Errorf("ndr: field %s: %w", f.Name, err)
			}
			continue
		}
		if err := marshalFieldInline(e, rv.Field(i), tag, deferred, embedded); err != nil {
			return -1, fmt.Errorf("ndr: field %s: %w", f.Name, err)
		}
	}
	return retvalIdx, nil
}

// runDeferred executes each deferred referent body in order. An index loop is used so
// that a referent body which itself defers further referents (a pointer reached through
// another pointer) is processed in the same pass, matching NDR's depth-first referent
// traversal ([C706] section 14.3.10).
func runDeferred(deferred []func() error) error {
	for i := 0; i < len(deferred); i++ {
		if err := deferred[i](); err != nil {
			return err
		}
	}
	return nil
}

// marshalFieldInline writes a field's inline representation, enqueueing the referent
// body for pointers.
func marshalFieldInline(e *Encoder, fv reflect.Value, tag fieldTag, deferred *[]func() error, embedded bool) error {
	if tag.align > 0 {
		e.Align(tag.align)
	}
	if m, ok := asMarshaler(fv); ok {
		e.Align(m.AlignmentNDR())
		return m.MarshalNDR(e)
	}

	if tag.pipe {
		return marshalPipe(e, fv, tag)
	}

	if isPointerLike(fv, tag) {
		kind := tag.ptr
		if kind == ptrNone {
			kind = ptrUnique // a bare Go pointer defaults to [unique]
		}
		if kind == ptrRef {
			if isNilReferent(fv) {
				return fmt.Errorf("ndr: nil [ref] pointer")
			}
			// A top-level [ref] parameter has no referent id and its referent is
			// marshalled in place. An embedded [ref] pointer is represented by a
			// referent-id placeholder (4 octets under NDR20, 8 under NDR64) with its
			// referent deferred, like other embedded pointers ([C706] section 14.3.10).
			if !embedded {
				return marshalReferentBody(e, fv, tag)
			}
			e.writeReferent(e.nextReferent())
			body := fv
			*deferred = append(*deferred, func() error { return marshalReferentBody(e, body, tag) })
			return nil
		}
		if isNilReferent(fv) {
			e.writeReferent(0) // NULL referent
			return nil
		}
		e.writeReferent(e.nextReferent())
		// A top-level [unique]/[full] pointer parameter marshals its referent in place,
		// immediately after the referent id; only an embedded pointer defers the body to
		// the end of the enclosing construction ([C706] section 14.3.10, and observed on
		// the wire for LSA [out] parameters such as LsarLookupSids ReferencedDomains).
		if !embedded {
			return marshalReferentBody(e, fv, tag)
		}
		body := fv
		*deferred = append(*deferred, func() error { return marshalReferentBody(e, body, tag) })
		return nil
	}

	return marshalInlineValue(e, fv, tag, deferred, embedded)
}

// marshalReferentBody marshals the representation a pointer refers to.
func marshalReferentBody(e *Encoder, fv reflect.Value, tag fieldTag) error {
	if fv.Kind() == reflect.Pointer {
		fv = fv.Elem()
	}
	return marshalInlineValue(e, fv, tag, nil, true)
}

// marshalInlineValue marshals a non-pointer value in place. deferred may be nil when
// the value is itself a referent body (its own pointers create a fresh construction
// scope via marshalStruct). embedded is true when the value is nested inside another
// construction rather than being a top-level parameter.
func marshalInlineValue(e *Encoder, fv reflect.Value, tag fieldTag, deferred *[]func() error, embedded bool) error {
	if m, ok := asMarshaler(fv); ok {
		e.Align(m.AlignmentNDR())
		return m.MarshalNDR(e)
	}

	if tag.isEnum {
		return marshalEnum(e, fv)
	}

	if isString, wide := stringMode(fv, tag); isString {
		if wide {
			e.writeWString(fv.String())
		} else {
			e.writeAString(fv.String())
		}
		return nil
	}

	switch fv.Kind() {
	case reflect.Struct:
		if swIdx := unionSwitchIndex(fv.Type()); swIdx >= 0 {
			return marshalUnion(e, fv, swIdx)
		}
		// A struct value nested inside another construction (embedded) is part of that
		// construction: its pointers defer to the enclosing flush point, not the end of
		// this struct ([C706] 14.3.12.3 — deferral iterates to the outermost construction).
		// A top-level parameter struct, or a referent body (deferred == nil), is its own
		// construction with its own deferral scope.
		if embedded && deferred != nil {
			_, err := marshalStructFields(e, fv, true, deferred)
			return err
		}
		return marshalStruct(e, fv, true)
	case reflect.Slice:
		// maximum_count is the element count, unless size_is names a literal bound
		// (e.g. MS-SAMR [size_is(1000)]), in which case the server requires that exact
		// constant on the wire even when fewer elements are actually sent.
		maxCount := uint32(fv.Len())
		if tag.sizeConstSet {
			maxCount = tag.sizeConst
		}
		if tag.varying {
			// actual_count is the slice length unless length_is fixes it (e.g.
			// RPC_UNICODE_STRING's Length/2), in which case only that many leading
			// elements are the valid, transmitted portion.
			actualCount := uint32(fv.Len())
			if tag.lengthConstSet {
				actualCount = tag.lengthConst
			}
			return marshalConformantVaryingArray(e, fv, maxCount, actualCount, elementTag(tag))
		}
		return marshalConformantArray(e, fv, maxCount, elementTag(tag))
	case reflect.Array:
		if fv.Type().Elem().Kind() == reflect.Uint8 {
			b := make([]byte, fv.Len())
			reflect.Copy(reflect.ValueOf(b), fv)
			e.WriteBytes(b)
			return nil
		}
		return fmt.Errorf("ndr: unsupported array element %s", fv.Type().Elem().Kind())
	default:
		return marshalScalar(e, fv)
	}
}

// marshalEnum writes an integer-kinded value as an NDR enum (2 octets under NDR20, 4
// under NDR64), regardless of the Go integer width. An NDR enum is always a 16-bit
// value on the wire under NDR20.
func marshalEnum(e *Encoder, fv reflect.Value) error {
	switch {
	case isUintKind(fv.Kind()):
		e.writeEnum(fv.Uint())
	case fv.CanInt():
		e.writeEnum(uint64(fv.Int()))
	default:
		return fmt.Errorf("ndr: enum field must be an integer, got %s", fv.Kind())
	}
	return nil
}

// unmarshalEnum reads an NDR enum into an integer-kinded value.
func unmarshalEnum(d *Decoder, fv reflect.Value) error {
	v, err := d.readEnum()
	if err != nil {
		return err
	}
	if isUintKind(fv.Kind()) {
		fv.SetUint(v)
	} else {
		fv.SetInt(int64(v))
	}
	return nil
}

// marshalScalar writes an integer or boolean primitive with NDR alignment.
func marshalScalar(e *Encoder, fv reflect.Value) error {
	switch fv.Kind() {
	case reflect.Bool:
		if fv.Bool() {
			e.WriteUint8(1)
		} else {
			e.WriteUint8(0)
		}
	case reflect.Uint8:
		e.WriteUint8(uint8(fv.Uint()))
	case reflect.Uint16:
		e.WriteUint16(uint16(fv.Uint()))
	case reflect.Uint32, reflect.Uint:
		e.WriteUint32(uint32(fv.Uint()))
	case reflect.Uint64:
		e.WriteUint64(fv.Uint())
	case reflect.Int8:
		e.WriteUint8(uint8(fv.Int()))
	case reflect.Int16:
		e.WriteUint16(uint16(fv.Int()))
	case reflect.Int32, reflect.Int:
		e.WriteUint32(uint32(fv.Int()))
	case reflect.Int64:
		e.WriteUint64(uint64(fv.Int()))
	default:
		return fmt.Errorf("ndr: unsupported scalar kind %s", fv.Kind())
	}
	return nil
}

// ---- unmarshalling --------------------------------------------------------------

func unmarshalStruct(d *Decoder, rv reflect.Value, embedded bool) error {
	var deferred []func() error
	retvalIdx, err := unmarshalStructFields(d, rv, embedded, &deferred)
	if err != nil {
		return err
	}
	if err := runDeferred(deferred); err != nil {
		return err
	}
	// The RPC return value follows all [out] parameters and their deferred referents.
	if retvalIdx >= 0 {
		rt := rv.Type()
		f := rt.Field(retvalIdx)
		var rdef []func() error
		if err := unmarshalFieldInline(d, rv.Field(retvalIdx), parseTag(f.Tag.Get("ndr")), &rdef, embedded); err != nil {
			return fmt.Errorf("ndr: field %s: %w", f.Name, err)
		}
		if err := runDeferred(rdef); err != nil {
			return err
		}
	}
	return nil
}

// unmarshalStructFields reads a struct's inline representation, appending each deferred
// referent body to *deferred rather than reading it. The caller owns the flush; see
// marshalStructFields for why arrays need this. Returns the `retval` field index or -1.
func unmarshalStructFields(d *Decoder, rv reflect.Value, embedded bool, deferred *[]func() error) (int, error) {
	rt := rv.Type()
	confIdx := embeddedConformantIndex(rt, rv)
	d.Align(leadingAlignment(rt, confIdx, d.syntax)) // structure aligned to its largest member ([C706] 14.2.2)
	var confCount uint64
	if confIdx >= 0 {
		c, err := d.readCount() // hoisted maximum_count (determinant-aligned; elements self-align)
		if err != nil {
			return -1, err
		}
		confCount = c
		// After the determinant, the structure body begins at the structure's natural
		// alignment (8 for a conformant array of 8-octet-aligned elements). ndrAlignment
		// of the struct under-reports this (a slice field counts as 4), so align to the
		// element's alignment directly. See the encode side; without this the member
		// after the count is read 4-aligned and every record drifts.
		d.Align(ndrAlignment(rv.Field(confIdx).Type().Elem(), d.syntax))
	}
	retvalIdx := retvalIndex(rt)
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tag := parseTag(f.Tag.Get("ndr"))
		if tag.skip {
			continue
		}
		if i == retvalIdx {
			continue // the RPC return value is decoded last, after deferred referents
		}
		if i == confIdx {
			n := int(confCount)
			if tag.varying {
				if _, err := d.readCount(); err != nil { // offset
					return -1, fmt.Errorf("ndr: field %s: %w", f.Name, err)
				}
				actual, err := d.readCount() // actual_count
				if err != nil {
					return -1, fmt.Errorf("ndr: field %s: %w", f.Name, err)
				}
				n = int(actual)
			}
			if err := unmarshalElements(d, rv.Field(i), n, elementTag(tag)); err != nil {
				return -1, fmt.Errorf("ndr: field %s: %w", f.Name, err)
			}
			continue
		}
		if err := unmarshalFieldInline(d, rv.Field(i), tag, deferred, embedded); err != nil {
			return -1, fmt.Errorf("ndr: field %s: %w", f.Name, err)
		}
	}
	return retvalIdx, nil
}

func unmarshalFieldInline(d *Decoder, fv reflect.Value, tag fieldTag, deferred *[]func() error, embedded bool) error {
	if tag.align > 0 {
		d.Align(tag.align)
	}
	if m, ok := asMarshalerAddr(fv); ok {
		d.Align(m.AlignmentNDR())
		return m.UnmarshalNDR(d)
	}

	if tag.pipe {
		return unmarshalPipe(d, fv, tag)
	}

	if isPointerLike(fv, tag) {
		kind := tag.ptr
		if kind == ptrNone {
			kind = ptrUnique
		}
		if kind == ptrRef {
			// Top-level [ref]: referent in place. Embedded [ref]: a referent-id
			// placeholder (4 octets under NDR20, 8 under NDR64) then a deferred referent.
			if !embedded {
				return unmarshalReferentBody(d, fv, tag)
			}
			if _, err := d.readReferent(); err != nil {
				return err
			}
			target := fv
			*deferred = append(*deferred, func() error { return unmarshalReferentBody(d, target, tag) })
			return nil
		}
		refid, err := d.readReferent()
		if err != nil {
			return err
		}
		if refid == 0 {
			return nil // NULL: leave the zero value
		}
		// Symmetric with marshal: a top-level [unique]/[full] pointer's referent is read
		// in place, right after the referent id; only an embedded pointer's body is
		// deferred to the end of the enclosing construction.
		if !embedded {
			return unmarshalReferentBody(d, fv, tag)
		}
		target := fv
		*deferred = append(*deferred, func() error { return unmarshalReferentBody(d, target, tag) })
		return nil
	}

	return unmarshalInlineValue(d, fv, tag, deferred, embedded)
}

func unmarshalReferentBody(d *Decoder, fv reflect.Value, tag fieldTag) error {
	if fv.Kind() == reflect.Pointer {
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		fv = fv.Elem()
	}
	return unmarshalInlineValue(d, fv, tag, nil, true) // a referent body is a fresh construction
}

func unmarshalInlineValue(d *Decoder, fv reflect.Value, tag fieldTag, deferred *[]func() error, embedded bool) error {
	if m, ok := asMarshalerAddr(fv); ok {
		d.Align(m.AlignmentNDR())
		return m.UnmarshalNDR(d)
	}

	if tag.isEnum {
		return unmarshalEnum(d, fv)
	}

	if isString, wide := stringMode(fv, tag); isString {
		var s string
		var err error
		if wide {
			s, err = d.readWString()
		} else {
			s, err = d.readAString()
		}
		if err != nil {
			return err
		}
		fv.SetString(s)
		return nil
	}

	switch fv.Kind() {
	case reflect.Struct:
		if swIdx := unionSwitchIndex(fv.Type()); swIdx >= 0 {
			return unmarshalUnion(d, fv, swIdx)
		}
		// Symmetric with marshal: a struct value nested inside another construction threads
		// the enclosing deferred queue; a top-level parameter struct or a referent body
		// (deferred == nil) starts its own construction.
		if embedded && deferred != nil {
			_, err := unmarshalStructFields(d, fv, true, deferred)
			return err
		}
		return unmarshalStruct(d, fv, true)
	case reflect.Slice:
		if tag.varying {
			return unmarshalConformantVaryingArray(d, fv, elementTag(tag))
		}
		return unmarshalConformantArray(d, fv, elementTag(tag))
	case reflect.Array:
		if fv.Type().Elem().Kind() == reflect.Uint8 {
			b, err := d.ReadBytes(fv.Len())
			if err != nil {
				return err
			}
			reflect.Copy(fv, reflect.ValueOf(b))
			return nil
		}
		return fmt.Errorf("ndr: unsupported array element %s", fv.Type().Elem().Kind())
	default:
		return unmarshalScalar(d, fv)
	}
}

func unmarshalScalar(d *Decoder, fv reflect.Value) error {
	switch fv.Kind() {
	case reflect.Bool:
		v, err := d.ReadUint8()
		if err != nil {
			return err
		}
		fv.SetBool(v != 0)
	case reflect.Uint8:
		v, err := d.ReadUint8()
		if err != nil {
			return err
		}
		fv.SetUint(uint64(v))
	case reflect.Uint16:
		v, err := d.ReadUint16()
		if err != nil {
			return err
		}
		fv.SetUint(uint64(v))
	case reflect.Uint32, reflect.Uint:
		v, err := d.ReadUint32()
		if err != nil {
			return err
		}
		fv.SetUint(uint64(v))
	case reflect.Uint64:
		v, err := d.ReadUint64()
		if err != nil {
			return err
		}
		fv.SetUint(v)
	case reflect.Int8:
		v, err := d.ReadUint8()
		if err != nil {
			return err
		}
		fv.SetInt(int64(int8(v)))
	case reflect.Int16:
		v, err := d.ReadUint16()
		if err != nil {
			return err
		}
		fv.SetInt(int64(int16(v)))
	case reflect.Int32, reflect.Int:
		v, err := d.ReadUint32()
		if err != nil {
			return err
		}
		fv.SetInt(int64(int32(v)))
	case reflect.Int64:
		v, err := d.ReadUint64()
		if err != nil {
			return err
		}
		fv.SetInt(int64(v))
	default:
		return fmt.Errorf("ndr: unsupported scalar kind %s", fv.Kind())
	}
	return nil
}

// ---- helpers --------------------------------------------------------------------

// embeddedConformantIndex returns the field index of an embedded (non-pointer)
// conformant array member, or -1 if there is none. NDR permits at most one such
// member per structure (the last one); if several match, the last is used.
func embeddedConformantIndex(rt reflect.Type, rv reflect.Value) int {
	idx := -1
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tag := parseTag(f.Tag.Get("ndr"))
		if tag.skip {
			continue
		}
		if isEmbeddedConformantArray(rv.Field(i), tag) {
			idx = i
		}
	}
	return idx
}

// countTargets maps the name of each field that is named by another field's
// size_is/length_is to the element count of that array. The count is derived from the
// array's length so the transmitted count field always matches the elements ([C706]
// section 14.3.3.1; [MS-RPCE] Conformant Arrays).
func countTargets(rt reflect.Type, rv reflect.Value) map[string]uint64 {
	m := map[string]uint64{}
	for j := 0; j < rt.NumField(); j++ {
		f := rt.Field(j)
		if f.PkgPath != "" {
			continue
		}
		tag := parseTag(f.Tag.Get("ndr"))
		if tag.skip || rv.Field(j).Kind() != reflect.Slice {
			continue
		}
		n := uint64(rv.Field(j).Len())
		// A divisor form (size_is(Field/N)) names a field in different units than the
		// elements — its value is not the element count, so it is resolved separately by
		// resolveSiblingCounts and must not be overwritten with the slice length here.
		//
		// Independent bounds (a varying array whose size_is and length_is name *different*
		// sibling fields — e.g. RPC_SECURITY_DESCRIPTOR's CbIn/CbOut, LSAPR_CR_CIPHER_VALUE's
		// MaximumLength/Length) are likewise caller-owned: the maximum_count (capacity) and
		// the actual_count (valid length) genuinely differ, so neither may be clobbered with
		// the slice length. They are resolved from the field values in resolveSiblingCounts.
		if tag.sizeIs != "" && tag.sizeDiv == 0 && !independentVaryingBounds(tag) {
			m[tag.sizeIs] = n
		}
		if tag.lengthIs != "" && tag.lengthDiv == 0 && !independentVaryingBounds(tag) {
			m[tag.lengthIs] = n
		}
	}
	return m
}

// resolveSiblingCounts rewrites a field's size_is/length_is divisor form
// (size_is(Field/N), length_is(Field/N)) into literal maximum_count/actual_count values
// by reading the named sibling field from rv and dividing by N. This carries count
// fields expressed in different units than the array elements — notably
// RPC_UNICODE_STRING, whose Length/MaximumLength are byte counts driving a wchar array
// ([MS-DTYP] 2.3.10). Plain sibling size_is/length_is (no divisor) are left untouched;
// their counts are derived from the array length by countTargets.
func resolveSiblingCounts(tag *fieldTag, rt reflect.Type, rv reflect.Value) {
	// Independent bounds: a varying array whose size_is and length_is name distinct
	// sibling fields. The maximum_count (capacity) and actual_count (valid length) are
	// taken verbatim from those fields rather than from the slice length, so a caller can
	// transmit fewer valid elements than the buffer's capacity — the case the key-security
	// RPC_SECURITY_DESCRIPTOR (CbIn/CbOut) and LSAPR_CR_CIPHER_VALUE (MaximumLength/Length)
	// require ([MS-RRP] 2.2.8, [MS-LSAD] 2.2.6.1).
	if independentVaryingBounds(*tag) {
		if v, ok := siblingUint(rt, rv, tag.sizeIs); ok {
			tag.sizeConst = uint32(v)
			tag.sizeConstSet = true
		}
		if v, ok := siblingUint(rt, rv, tag.lengthIs); ok {
			tag.lengthConst = uint32(v)
			tag.lengthConstSet = true
		}
		return
	}
	if tag.sizeIs != "" && tag.sizeDiv > 0 {
		if v, ok := siblingUint(rt, rv, tag.sizeIs); ok {
			tag.sizeConst = uint32(v / uint64(tag.sizeDiv))
			tag.sizeConstSet = true
		}
	}
	if tag.lengthIs != "" && tag.lengthDiv > 0 {
		if v, ok := siblingUint(rt, rv, tag.lengthIs); ok {
			tag.lengthConst = uint32(v / uint64(tag.lengthDiv))
			tag.lengthConstSet = true
		}
	}
}

// independentVaryingBounds reports whether an array's size_is and length_is name two
// *different* plain sibling fields (no divisor): the conformant-varying case where the
// maximum_count (capacity) and the actual_count (valid length) are independent values the
// caller owns, rather than both being derived from the slice length. Divisor forms
// (size_is(Field/N)) and the common size_is==length_is form are excluded.
func independentVaryingBounds(tag fieldTag) bool {
	return tag.varying &&
		tag.sizeIs != "" && tag.lengthIs != "" &&
		tag.sizeIs != tag.lengthIs &&
		tag.sizeDiv == 0 && tag.lengthDiv == 0
}

// siblingUint reads the unsigned integer value of the named field of struct value rv,
// reporting false if there is no such field or it is not an unsigned integer kind.
func siblingUint(rt reflect.Type, rv reflect.Value, name string) (uint64, bool) {
	f, ok := rt.FieldByName(name)
	if !ok || len(f.Index) != 1 {
		return 0, false
	}
	fv := rv.Field(f.Index[0])
	if !isUintKind(fv.Kind()) {
		return 0, false
	}
	return fv.Uint(), true
}

// retvalIndex returns the field index tagged `ndr:"retval"` (the RPC return value),
// or -1. NDR places the return value after all [out] parameters and their deferred
// referents, so the walker handles it separately from the inline fields.
func retvalIndex(rt reflect.Type) int {
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.PkgPath != "" {
			continue
		}
		if parseTag(f.Tag.Get("ndr")).retval {
			return i
		}
	}
	return -1
}

func isUintKind(k reflect.Kind) bool {
	switch k {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

// isEmbeddedConformantArray reports whether fv is an inline (non-pointer) conformant
// array, i.e. a slice that is not itself behind a pointer. A slice behind a pointer is
// a referent, not an embedded array, and is not hoisted.
func isEmbeddedConformantArray(fv reflect.Value, tag fieldTag) bool {
	// A pipe is a chunked stream, not a conformant array, so its count must not be
	// hoisted to the front of the enclosing struct ([C706] 14.7).
	//
	// An "inline"-tagged slice is a top-level conformant[-varying] array whose
	// maximum_count is transmitted in place, immediately before the array, rather than
	// hoisted to the front of the enclosing struct. This is the form of a bare,
	// non-pointer, top-level [out] array parameter such as ept_lookup's entries[]
	// ([C706] Appendix O; [MS-RPCE] 2.2.1.2.4) — the structure-embedded hoist rule
	// ([MS-RPCE] "Structure Containing a Conformant Varying Array") does not apply, so it
	// must not be treated as an embedded conformant array.
	return fv.Kind() == reflect.Slice && !tag.pipe && !tag.inline && !isPointerLike(fv, tag)
}

// ndrAlignment returns the NDR alignment, in octets, of a value of type t under the
// given transfer syntax. Strings, pointers, and slices lead with a referent id or a
// conformance count, which widen from 4 octets 4-aligned (NDR20) to 8 octets 8-aligned
// (NDR64), so those kinds align to 8 under NDR64 ([MS-RPCE] section 2.2.5). Fixed scalar
// kinds keep their natural alignment in both syntaxes.
func ndrAlignment(t reflect.Type, syntax Syntax) int {
	if reflect.PointerTo(t).Implements(marshalerType) {
		return reflect.New(t).Interface().(Marshaler).AlignmentNDR()
	}
	switch t.Kind() {
	case reflect.Bool, reflect.Uint8, reflect.Int8:
		return 1
	case reflect.Uint16, reflect.Int16:
		return 2
	case reflect.Uint32, reflect.Int32, reflect.Uint, reflect.Int:
		return 4
	case reflect.Uint64, reflect.Int64:
		return 8
	case reflect.String, reflect.Pointer, reflect.Slice:
		// A string/slice begins with a conformance count and a pointer with a referent
		// id; both are 8-octet 8-aligned under NDR64 and 4-octet 4-aligned under NDR20.
		if syntax == NDR64 {
			return 8
		}
		return 4
	case reflect.Array:
		// A fixed array aligns to its element. Under NDR20 the historical behaviour is
		// the catch-all alignment of 4, preserved to keep NDR20 output byte-identical.
		if syntax == NDR64 {
			return ndrAlignment(t.Elem(), syntax)
		}
		return 4
	case reflect.Struct:
		a := 1
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue
			}
			// NDR20 aligns to the widest member type, ignoring tags (preserved exactly).
			// NDR64 is tag-aware: a value field tagged as a pointer ([unique]/[ref]/[ptr])
			// is a referent id inline, so it aligns to 8 regardless of the value's layout.
			var fa int
			if syntax == NDR64 {
				tag := parseTag(f.Tag.Get("ndr"))
				if tag.skip {
					continue
				}
				fa = fieldNDRAlignment(f.Type, tag, syntax)
			} else {
				fa = ndrAlignment(f.Type, syntax)
			}
			if fa > a {
				a = fa
			}
		}
		return a
	default:
		return 4
	}
}

// fieldNDRAlignment returns the alignment of a struct field, accounting for tags whose
// wire representation differs from the field's Go layout: a pointer tag on a non-pointer
// type (a referent id) and an enum (2 octets under NDR20, 4 under NDR64).
func fieldNDRAlignment(t reflect.Type, tag fieldTag, syntax Syntax) int {
	if tag.isEnum {
		if syntax == NDR64 {
			return 4
		}
		return 2
	}
	if tag.ptr != ptrNone && t.Kind() != reflect.Pointer {
		if syntax == NDR64 {
			return 8
		}
		return 4
	}
	return ndrAlignment(t, syntax)
}

// leadingAlignment returns the alignment applied at the start of a structure's
// representation. For a plain structure this is its natural alignment (the largest
// member's, [C706] 14.2.2). For a structure with an embedded conformant array (confIdx >=
// 0) the representation begins with the hoisted maximum_count size determinant, which is
// aligned to the determinant width (4 under NDR20, 8 under NDR64) and the alignment of the
// NON-array members — but NOT to the array element's alignment. The array elements are
// aligned to their own type when written/read, so a conformant array of 8-octet-aligned
// elements (e.g. one whose elements contain a hyper/int64) does not 8-align the
// maximum_count. This matches observed Windows behaviour (e.g. DS_REPL_CURSORS); inflating
// the determinant alignment to the element alignment shifts every following field.
func leadingAlignment(rt reflect.Type, confIdx int, syntax Syntax) int {
	if confIdx < 0 {
		return ndrAlignment(rt, syntax)
	}
	a := 4
	if syntax == NDR64 {
		a = 8
	}
	for i := 0; i < rt.NumField(); i++ {
		if i == confIdx {
			continue
		}
		f := rt.Field(i)
		if f.PkgPath != "" {
			continue
		}
		if fa := fieldNDRAlignment(f.Type, parseTag(f.Tag.Get("ndr")), syntax); fa > a {
			a = fa
		}
	}
	return a
}

// countAlignment is the alignment of an NDR size determinant (maximum_count, offset,
// actual_count): 4 octets under NDR20, 8 under NDR64 ([MS-RPCE] 2.2.5). The determinant
// is aligned to its own width, NOT to the array element alignment — the elements
// element-align themselves when written/read. Aligning the determinant to an 8-octet
// element alignment (an array of structs containing a hyper/int64, e.g. REPLVALINF_V1 or
// PROPERTY_META_DATA_EXT) would shift it and corrupt the parse. This is the
// standalone-array counterpart of leadingAlignment's fix for the embedded (hoisted) case.
func countAlignment(syntax Syntax) int {
	if syntax == NDR64 {
		return 8
	}
	return 4
}

// marshalConformantArray writes a conformant array: a maximum_count followed by the
// elements, per [MS-RPCE] Conformant Arrays:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/140b01a3-979b-43af-b1e3-28f248db8f03
func marshalConformantArray(e *Encoder, slice reflect.Value, maxCount uint32, elemTag fieldTag) error {
	e.Align(countAlignment(e.syntax))
	e.writeCount(uint64(maxCount))
	return marshalElements(e, slice, elemTag)
}

// unmarshalConformantArray reads a maximum_count-prefixed array into slice.
func unmarshalConformantArray(d *Decoder, slice reflect.Value, elemTag fieldTag) error {
	d.Align(countAlignment(d.syntax))
	n, err := d.readCount()
	if err != nil {
		return err
	}
	return unmarshalElements(d, slice, int(n), elemTag)
}

// marshalConformantVaryingArray writes a conformant-varying array: maximum_count,
// offset, actual_count, then the elements, per [MS-RPCE] Conformant Varying Arrays:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/3acb31b0-b873-4aaf-8503-9727ec40fbec
// Only the first actualCount elements are transmitted (offset 0). actualCount is the
// slice length unless length_is fixed it (RPC_UNICODE_STRING's Length/2), and maxCount
// is the slice length unless a size_is bound was given, in which case those values are
// sent while only the valid leading elements follow.
func marshalConformantVaryingArray(e *Encoder, slice reflect.Value, maxCount, actualCount uint32, elemTag fieldTag) error {
	if int(actualCount) > slice.Len() {
		actualCount = uint32(slice.Len())
	}
	e.Align(countAlignment(e.syntax))
	e.writeCount(uint64(maxCount))    // maximum_count
	e.writeCount(0)                   // offset
	e.writeCount(uint64(actualCount)) // actual_count
	return marshalElements(e, slice.Slice(0, int(actualCount)), elemTag)
}

// unmarshalConformantVaryingArray reads a conformant-varying array, using the
// actual_count to size the result.
func unmarshalConformantVaryingArray(d *Decoder, slice reflect.Value, elemTag fieldTag) error {
	d.Align(countAlignment(d.syntax))
	if _, err := d.readCount(); err != nil { // maximum_count
		return err
	}
	if _, err := d.readCount(); err != nil { // offset
		return err
	}
	actual, err := d.readCount() // actual_count
	if err != nil {
		return err
	}
	return unmarshalElements(d, slice, int(actual), elemTag)
}

// marshalElements writes the elements of a slice with no count prefix, in two passes:
// every element's fixed (inline) part first, then every element's deferred referents,
// both in element order. NDR defers the referents of pointers embedded in an array to
// a position after the entire array body, so an array of pointers, or of structs that
// themselves contain pointers, cannot be encoded by recursing fully into each element
// in turn ([C706] section 14.3.10).
// marshalPipe writes an NDR pipe ([C706] section 14.7): a sequence of chunks terminated
// by an empty chunk (a count of 0). The whole slice is transmitted as a single chunk. A
// pipe of a scalar (e.g. EFS_EXIM_PIPE = pipe of bytes) goes through marshalElements,
// which has a byte fast path.
//
// The chunk count widens with the syntax (4 octets under NDR20, 8 under NDR64) like any
// NDR count. Under NDR64, each chunk's elements are additionally followed by the two's-
// complement negation of the count (an 8-octet trailer): [+count][elements][-count],
// then a 0 terminator. This bracketing is NDR64-specific — verified on the wire against a
// Windows Server 2016 EfsRpcReadFileRaw response, whose chunks decoded identically under
// both syntaxes (see network/dcerpc/ndr/ndr64_pipe_test.go).
func marshalPipe(e *Encoder, fv reflect.Value, tag fieldTag) error {
	if fv.Kind() != reflect.Slice {
		return fmt.Errorf("ndr: pipe field must be a slice, got %s", fv.Kind())
	}
	if n := fv.Len(); n > 0 {
		e.writeCount(uint64(n))
		if err := marshalElements(e, fv, elementTag(tag)); err != nil {
			return err
		}
		if e.syntax == NDR64 {
			e.WriteUint64(^uint64(n) + 1) // -count trailer
		}
	}
	e.writeCount(0) // terminating empty chunk
	return nil
}

// unmarshalPipe reads an NDR pipe: chunks of (count, count elements) until a 0-count
// chunk, concatenating the chunks into the slice. Under NDR64 it also consumes the
// 8-octet -count trailer after each chunk's elements. unmarshalElements bounds each
// chunk's count against the remaining input, so a hostile count cannot over-allocate.
func unmarshalPipe(d *Decoder, fv reflect.Value, tag fieldTag) error {
	if fv.Kind() != reflect.Slice {
		return fmt.Errorf("ndr: pipe field must be a slice, got %s", fv.Kind())
	}
	acc := reflect.MakeSlice(fv.Type(), 0, 0)
	for {
		count, err := d.readCount()
		if err != nil {
			return err
		}
		if count == 0 {
			break
		}
		chunk := reflect.New(fv.Type()).Elem()
		if err := unmarshalElements(d, chunk, int(count), elementTag(tag)); err != nil {
			return err
		}
		acc = reflect.AppendSlice(acc, chunk)
		if d.syntax == NDR64 {
			if _, err := d.ReadUint64(); err != nil { // -count trailer
				return err
			}
		}
	}
	fv.Set(acc)
	return nil
}

func marshalElements(e *Encoder, slice reflect.Value, elemTag fieldTag) error {
	if slice.Type().Elem().Kind() == reflect.Uint8 {
		e.WriteBytes(slice.Bytes()) // fast path for byte arrays
		return nil
	}
	var deferred []func() error
	for i := 0; i < slice.Len(); i++ {
		if err := marshalElement(e, slice.Index(i), elemTag, &deferred); err != nil {
			return fmt.Errorf("element %d: %w", i, err)
		}
	}
	return runDeferred(deferred)
}

// marshalElement writes one array element's inline part, appending its deferred
// referents (if any) to *deferred. Pointer elements and struct elements that contain
// pointers route their referents up to the array's shared queue so they land after the
// whole array; pointer-free elements have no deferral and are written in place.
func marshalElement(e *Encoder, ev reflect.Value, tag fieldTag, deferred *[]func() error) error {
	if isPointerLike(ev, tag) {
		kind := tag.ptr
		if kind == ptrNone {
			kind = ptrUnique // a bare Go pointer element defaults to [unique]
		}
		// An array of [ref] pointers transmits no referent id per element ([C706]
		// section 14.3.10: "the special case of an array of reference pointers ... has
		// no NDR representation"); only the referent bodies are deferred. [unique]/[ptr]
		// elements carry a referent id (0 for NULL) inline, body deferred.
		if kind == ptrRef {
			if isNilReferent(ev) {
				return fmt.Errorf("ndr: nil [ref] array element")
			}
		} else {
			if isNilReferent(ev) {
				e.writeReferent(0) // NULL referent
				return nil
			}
			e.writeReferent(e.nextReferent())
		}
		body := ev
		*deferred = append(*deferred, func() error { return marshalReferentBody(e, body, tag) })
		return nil
	}
	// A struct element defers its own embedded pointers into the array's queue rather
	// than flushing them after itself, so they too land after the whole array.
	if ev.Kind() == reflect.Struct {
		if _, ok := asMarshaler(ev); !ok {
			if swIdx := unionSwitchIndex(ev.Type()); swIdx >= 0 {
				return marshalUnion(e, ev, swIdx)
			}
			_, err := marshalStructFields(e, ev, true, deferred)
			return err
		}
	}
	return marshalInlineValue(e, ev, tag, nil, true)
}

// unmarshalElements reads n elements into slice (allocating a fresh backing array), in
// two passes mirroring marshalElements: every element's fixed part first, then every
// element's deferred referents, both in element order ([C706] section 14.3.10).
func unmarshalElements(d *Decoder, slice reflect.Value, n int, elemTag fieldTag) error {
	if n < 0 {
		return fmt.Errorf("ndr: negative element count %d", n)
	}
	// Every NDR element occupies at least one octet, so a count larger than the
	// remaining input cannot be valid. Reject it before allocating, otherwise an
	// attacker-controlled count would size an arbitrary allocation (OOM/panic).
	if n > d.Remaining() {
		return fmt.Errorf("ndr: element count %d exceeds %d remaining bytes", n, d.Remaining())
	}
	if slice.Type().Elem().Kind() == reflect.Uint8 {
		b, err := d.ReadBytes(n)
		if err != nil {
			return err
		}
		out := make([]byte, n)
		copy(out, b)
		slice.SetBytes(out)
		return nil
	}
	out := reflect.MakeSlice(slice.Type(), n, n)
	var deferred []func() error
	for i := 0; i < n; i++ {
		if err := unmarshalElement(d, out.Index(i), elemTag, &deferred); err != nil {
			return fmt.Errorf("element %d: %w", i, err)
		}
	}
	if err := runDeferred(deferred); err != nil {
		return err
	}
	slice.Set(out)
	return nil
}

// unmarshalElement reads one array element's inline part, appending its deferred
// referent reads (if any) to *deferred. It mirrors marshalElement: pointer elements
// read a referent id (none for a [ref] array) and defer the body; struct elements defer
// their embedded pointers into the array's queue; pointer-free elements read in place.
func unmarshalElement(d *Decoder, ev reflect.Value, tag fieldTag, deferred *[]func() error) error {
	if isPointerLike(ev, tag) {
		kind := tag.ptr
		if kind == ptrNone {
			kind = ptrUnique
		}
		if kind != ptrRef {
			refid, err := d.readReferent()
			if err != nil {
				return err
			}
			if refid == 0 {
				return nil // NULL: leave the zero value
			}
		}
		target := ev
		*deferred = append(*deferred, func() error { return unmarshalReferentBody(d, target, tag) })
		return nil
	}
	if ev.Kind() == reflect.Struct {
		if _, ok := asMarshalerAddr(ev); !ok {
			if swIdx := unionSwitchIndex(ev.Type()); swIdx >= 0 {
				return unmarshalUnion(d, ev, swIdx)
			}
			_, err := unmarshalStructFields(d, ev, true, deferred)
			return err
		}
	}
	return unmarshalInlineValue(d, ev, tag, nil, true)
}

// elementTag derives the tag applied to each element of an array from the array
// field's own tag. Only the element pointer attribute (`elem=ref|unique|ptr`) carries
// over; the array-level attributes (size_is, varying, …) describe the array, not its
// elements.
func elementTag(tag fieldTag) fieldTag {
	return fieldTag{ptr: tag.elemPtr}
}

func isPointerLike(fv reflect.Value, tag fieldTag) bool {
	if fv.Kind() == reflect.Pointer {
		return true
	}
	return tag.ptr != ptrNone
}

func isNilReferent(fv reflect.Value) bool {
	switch fv.Kind() {
	case reflect.Pointer, reflect.Slice:
		return fv.IsNil()
	default:
		return false // value strings are always present
	}
}

// stringMode reports whether fv is an NDR string and, if so, whether it is wide
// (UTF-16). The named aliases WSTR/STR select the mode by type; a plain string uses
// the tag and defaults to wide (the common MS-RPC case).
func stringMode(fv reflect.Value, tag fieldTag) (isString, wide bool) {
	switch {
	case fv.Type() == wstrType:
		return true, true
	case fv.Type() == strType:
		return true, false
	case fv.Kind() == reflect.String:
		return true, !tag.ascii
	default:
		return false, false
	}
}

func asMarshaler(fv reflect.Value) (Marshaler, bool) {
	if fv.Type().Implements(marshalerType) {
		return fv.Interface().(Marshaler), true
	}
	if fv.CanAddr() && fv.Addr().Type().Implements(marshalerType) {
		return fv.Addr().Interface().(Marshaler), true
	}
	return nil, false
}

func asMarshalerAddr(fv reflect.Value) (Marshaler, bool) {
	if fv.CanAddr() && fv.Addr().Type().Implements(marshalerType) {
		return fv.Addr().Interface().(Marshaler), true
	}
	if fv.Type().Implements(marshalerType) {
		return fv.Interface().(Marshaler), true
	}
	return nil, false
}

var marshalerType = reflect.TypeOf((*Marshaler)(nil)).Elem()

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
func Marshal(v any) ([]byte, error) {
	rv, err := structValue(reflect.ValueOf(v))
	if err != nil {
		return nil, err
	}
	e := NewEncoder()
	if err := marshalStruct(e, rv, false); err != nil {
		return nil, err
	}
	return e.Bytes(), nil
}

// Unmarshal decodes an NDR octet stream into v, which must be a non-nil pointer to a
// struct.
func Unmarshal(data []byte, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("ndr: Unmarshal requires a non-nil pointer, got %T", v)
	}
	sv, err := structValue(rv)
	if err != nil {
		return err
	}
	return unmarshalStruct(NewDecoder(data), sv, false)
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
	confIdx := embeddedConformantIndex(rt, rv)
	if confIdx >= 0 {
		e.Align(conformantArrayAlignment(rv.Field(confIdx).Type().Elem()))
		e.WriteUint32(uint32(rv.Field(confIdx).Len())) // hoisted maximum_count
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
				e.WriteUint32(0)                         // offset
				e.WriteUint32(uint32(rv.Field(i).Len())) // actual_count
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
			// 4-octet placeholder with its referent deferred, like other embedded
			// pointers ([C706] section 14.3.10).
			if !embedded {
				return marshalReferentBody(e, fv, tag)
			}
			e.WriteUint32(e.nextReferent())
			body := fv
			*deferred = append(*deferred, func() error { return marshalReferentBody(e, body, tag) })
			return nil
		}
		if isNilReferent(fv) {
			e.WriteUint32(0) // NULL referent
			return nil
		}
		e.WriteUint32(e.nextReferent())
		body := fv
		*deferred = append(*deferred, func() error { return marshalReferentBody(e, body, tag) })
		return nil
	}

	return marshalInlineValue(e, fv, tag, deferred)
}

// marshalReferentBody marshals the representation a pointer refers to.
func marshalReferentBody(e *Encoder, fv reflect.Value, tag fieldTag) error {
	if fv.Kind() == reflect.Pointer {
		fv = fv.Elem()
	}
	return marshalInlineValue(e, fv, tag, nil)
}

// marshalInlineValue marshals a non-pointer value in place. deferred may be nil when
// the value is itself a referent body (its own pointers create a fresh construction
// scope via marshalStruct).
func marshalInlineValue(e *Encoder, fv reflect.Value, tag fieldTag, deferred *[]func() error) error {
	if m, ok := asMarshaler(fv); ok {
		e.Align(m.AlignmentNDR())
		return m.MarshalNDR(e)
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
		return marshalStruct(e, fv, true) // members of any value reached here are embedded
	case reflect.Slice:
		if tag.varying {
			return marshalConformantVaryingArray(e, fv, elementTag(tag))
		}
		return marshalConformantArray(e, fv, elementTag(tag))
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
	var confCount uint32
	if confIdx >= 0 {
		d.Align(conformantArrayAlignment(rv.Field(confIdx).Type().Elem()))
		c, err := d.ReadUint32() // hoisted maximum_count
		if err != nil {
			return -1, err
		}
		confCount = c
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
				if _, err := d.ReadUint32(); err != nil { // offset
					return -1, fmt.Errorf("ndr: field %s: %w", f.Name, err)
				}
				actual, err := d.ReadUint32() // actual_count
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

	if isPointerLike(fv, tag) {
		kind := tag.ptr
		if kind == ptrNone {
			kind = ptrUnique
		}
		if kind == ptrRef {
			// Top-level [ref]: referent in place. Embedded [ref]: 4-octet placeholder
			// then a deferred referent.
			if !embedded {
				return unmarshalReferentBody(d, fv, tag)
			}
			if _, err := d.ReadUint32(); err != nil {
				return err
			}
			target := fv
			*deferred = append(*deferred, func() error { return unmarshalReferentBody(d, target, tag) })
			return nil
		}
		refid, err := d.ReadUint32()
		if err != nil {
			return err
		}
		if refid == 0 {
			return nil // NULL: leave the zero value
		}
		target := fv
		*deferred = append(*deferred, func() error { return unmarshalReferentBody(d, target, tag) })
		return nil
	}

	return unmarshalInlineValue(d, fv, tag)
}

func unmarshalReferentBody(d *Decoder, fv reflect.Value, tag fieldTag) error {
	if fv.Kind() == reflect.Pointer {
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		fv = fv.Elem()
	}
	return unmarshalInlineValue(d, fv, tag)
}

func unmarshalInlineValue(d *Decoder, fv reflect.Value, tag fieldTag) error {
	if m, ok := asMarshalerAddr(fv); ok {
		d.Align(m.AlignmentNDR())
		return m.UnmarshalNDR(d)
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
		if tag.sizeIs != "" {
			m[tag.sizeIs] = n
		}
		if tag.lengthIs != "" {
			m[tag.lengthIs] = n
		}
	}
	return m
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
	return fv.Kind() == reflect.Slice && !isPointerLike(fv, tag)
}

// ndrAlignment returns the NDR alignment, in octets, of a value of type t.
func ndrAlignment(t reflect.Type) int {
	if reflect.PointerTo(t).Implements(marshalerType) {
		return reflect.New(t).Interface().(Marshaler).AlignmentNDR()
	}
	switch t.Kind() {
	case reflect.Bool, reflect.Uint8, reflect.Int8:
		return 1
	case reflect.Uint16, reflect.Int16:
		return 2
	case reflect.Uint32, reflect.Int32, reflect.Uint, reflect.Int, reflect.String, reflect.Pointer:
		return 4
	case reflect.Uint64, reflect.Int64:
		return 8
	case reflect.Struct:
		a := 1
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).PkgPath != "" {
				continue
			}
			if fa := ndrAlignment(t.Field(i).Type); fa > a {
				a = fa
			}
		}
		return a
	default:
		return 4
	}
}

// conformantArrayAlignment returns the alignment of a conformant array of the given
// element type: the larger of the size determinant's alignment (4) and the element
// alignment ([C706] section 14.3.3.1).
func conformantArrayAlignment(elemType reflect.Type) int {
	if a := ndrAlignment(elemType); a > 4 {
		return a
	}
	return 4
}

// marshalConformantArray writes a conformant array: a maximum_count followed by the
// elements, per [MS-RPCE] Conformant Arrays:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/140b01a3-979b-43af-b1e3-28f248db8f03
func marshalConformantArray(e *Encoder, slice reflect.Value, elemTag fieldTag) error {
	e.Align(conformantArrayAlignment(slice.Type().Elem()))
	e.WriteUint32(uint32(slice.Len()))
	return marshalElements(e, slice, elemTag)
}

// unmarshalConformantArray reads a maximum_count-prefixed array into slice.
func unmarshalConformantArray(d *Decoder, slice reflect.Value, elemTag fieldTag) error {
	d.Align(conformantArrayAlignment(slice.Type().Elem()))
	n, err := d.ReadUint32()
	if err != nil {
		return err
	}
	return unmarshalElements(d, slice, int(n), elemTag)
}

// marshalConformantVaryingArray writes a conformant-varying array: maximum_count,
// offset, actual_count, then the elements, per [MS-RPCE] Conformant Varying Arrays:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/3acb31b0-b873-4aaf-8503-9727ec40fbec
// The full slice is transmitted (offset 0, actual_count == maximum_count == len).
func marshalConformantVaryingArray(e *Encoder, slice reflect.Value, elemTag fieldTag) error {
	e.Align(conformantArrayAlignment(slice.Type().Elem()))
	n := uint32(slice.Len())
	e.WriteUint32(n) // maximum_count
	e.WriteUint32(0) // offset
	e.WriteUint32(n) // actual_count
	return marshalElements(e, slice, elemTag)
}

// unmarshalConformantVaryingArray reads a conformant-varying array, using the
// actual_count to size the result.
func unmarshalConformantVaryingArray(d *Decoder, slice reflect.Value, elemTag fieldTag) error {
	d.Align(conformantArrayAlignment(slice.Type().Elem()))
	if _, err := d.ReadUint32(); err != nil { // maximum_count
		return err
	}
	if _, err := d.ReadUint32(); err != nil { // offset
		return err
	}
	actual, err := d.ReadUint32() // actual_count
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
				e.WriteUint32(0) // NULL referent
				return nil
			}
			e.WriteUint32(e.nextReferent())
		}
		body := ev
		*deferred = append(*deferred, func() error { return marshalReferentBody(e, body, tag) })
		return nil
	}
	// A struct element defers its own embedded pointers into the array's queue rather
	// than flushing them after itself, so they too land after the whole array.
	if ev.Kind() == reflect.Struct {
		if _, ok := asMarshaler(ev); !ok {
			_, err := marshalStructFields(e, ev, true, deferred)
			return err
		}
	}
	return marshalInlineValue(e, ev, tag, nil)
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
			refid, err := d.ReadUint32()
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
			_, err := unmarshalStructFields(d, ev, true, deferred)
			return err
		}
	}
	return unmarshalInlineValue(d, ev, tag)
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

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
	rt := rv.Type()
	confIdx := embeddedConformantIndex(rt, rv)
	if confIdx >= 0 {
		e.WriteUint32(uint32(rv.Field(confIdx).Len())) // hoisted maximum_count
	}
	// A field named by another field's size_is/length_is is the array's count; its
	// value is derived from the array length so the count and the elements cannot
	// disagree and the caller need not set it explicitly.
	counts := countTargets(rt, rv)
	var deferred []func() error
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.PkgPath != "" { // unexported
			continue
		}
		tag := parseTag(f.Tag.Get("ndr"))
		if tag.skip {
			continue
		}
		if c, ok := counts[f.Name]; ok && isUintKind(rv.Field(i).Kind()) {
			if tag.align > 0 {
				e.Align(tag.align)
			}
			tmp := reflect.New(f.Type).Elem()
			tmp.SetUint(c)
			if err := marshalScalar(e, tmp); err != nil {
				return fmt.Errorf("ndr: field %s: %w", f.Name, err)
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
			if err := marshalElements(e, rv.Field(i)); err != nil {
				return fmt.Errorf("ndr: field %s: %w", f.Name, err)
			}
			continue
		}
		if err := marshalFieldInline(e, rv.Field(i), tag, &deferred, embedded); err != nil {
			return fmt.Errorf("ndr: field %s: %w", f.Name, err)
		}
	}
	for _, fn := range deferred {
		if err := fn(); err != nil {
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
	if u, ok := asUnion(fv); ok {
		sw := u.SwitchValue()
		e.WriteUint32(sw) // encapsulated discriminant (4-octet, 4-aligned)
		return u.MarshalArm(e, sw)
	}
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
			return marshalConformantVaryingArray(e, fv)
		}
		return marshalConformantArray(e, fv)
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
	rt := rv.Type()
	confIdx := embeddedConformantIndex(rt, rv)
	var confCount uint32
	if confIdx >= 0 {
		c, err := d.ReadUint32() // hoisted maximum_count
		if err != nil {
			return err
		}
		confCount = c
	}
	var deferred []func() error
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tag := parseTag(f.Tag.Get("ndr"))
		if tag.skip {
			continue
		}
		if i == confIdx {
			n := int(confCount)
			if tag.varying {
				if _, err := d.ReadUint32(); err != nil { // offset
					return fmt.Errorf("ndr: field %s: %w", f.Name, err)
				}
				actual, err := d.ReadUint32() // actual_count
				if err != nil {
					return fmt.Errorf("ndr: field %s: %w", f.Name, err)
				}
				n = int(actual)
			}
			if err := unmarshalElements(d, rv.Field(i), n); err != nil {
				return fmt.Errorf("ndr: field %s: %w", f.Name, err)
			}
			continue
		}
		if err := unmarshalFieldInline(d, rv.Field(i), tag, &deferred, embedded); err != nil {
			return fmt.Errorf("ndr: field %s: %w", f.Name, err)
		}
	}
	for _, fn := range deferred {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
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
	if u, ok := asUnion(fv); ok {
		sw, err := d.ReadUint32()
		if err != nil {
			return err
		}
		return u.UnmarshalArm(d, sw)
	}
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
			return unmarshalConformantVaryingArray(d, fv)
		}
		return unmarshalConformantArray(d, fv)
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

// marshalConformantArray writes a conformant array: a maximum_count followed by the
// elements, per [MS-RPCE] Conformant Arrays:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/140b01a3-979b-43af-b1e3-28f248db8f03
func marshalConformantArray(e *Encoder, slice reflect.Value) error {
	e.WriteUint32(uint32(slice.Len()))
	return marshalElements(e, slice)
}

// unmarshalConformantArray reads a maximum_count-prefixed array into slice.
func unmarshalConformantArray(d *Decoder, slice reflect.Value) error {
	n, err := d.ReadUint32()
	if err != nil {
		return err
	}
	return unmarshalElements(d, slice, int(n))
}

// marshalConformantVaryingArray writes a conformant-varying array: maximum_count,
// offset, actual_count, then the elements, per [MS-RPCE] Conformant Varying Arrays:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/3acb31b0-b873-4aaf-8503-9727ec40fbec
// The full slice is transmitted (offset 0, actual_count == maximum_count == len).
func marshalConformantVaryingArray(e *Encoder, slice reflect.Value) error {
	n := uint32(slice.Len())
	e.WriteUint32(n) // maximum_count
	e.WriteUint32(0) // offset
	e.WriteUint32(n) // actual_count
	return marshalElements(e, slice)
}

// unmarshalConformantVaryingArray reads a conformant-varying array, using the
// actual_count to size the result.
func unmarshalConformantVaryingArray(d *Decoder, slice reflect.Value) error {
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
	return unmarshalElements(d, slice, int(actual))
}

// marshalElements writes the elements of a slice with no count prefix. Each element
// is marshalled via the inline-value path, so any supported element type (scalars,
// structs, Marshaler types) is handled with its natural alignment.
func marshalElements(e *Encoder, slice reflect.Value) error {
	if slice.Type().Elem().Kind() == reflect.Uint8 {
		e.WriteBytes(slice.Bytes()) // fast path for byte arrays
		return nil
	}
	for i := 0; i < slice.Len(); i++ {
		if err := marshalInlineValue(e, slice.Index(i), fieldTag{}, nil); err != nil {
			return fmt.Errorf("element %d: %w", i, err)
		}
	}
	return nil
}

// unmarshalElements reads n elements into slice (allocating a fresh backing array).
func unmarshalElements(d *Decoder, slice reflect.Value, n int) error {
	if n < 0 {
		return fmt.Errorf("ndr: negative element count %d", n)
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
	for i := 0; i < n; i++ {
		if err := unmarshalInlineValue(d, out.Index(i), fieldTag{}); err != nil {
			return fmt.Errorf("element %d: %w", i, err)
		}
	}
	slice.Set(out)
	return nil
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

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
	if err := marshalStruct(e, rv); err != nil {
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
	return unmarshalStruct(NewDecoder(data), sv)
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
func marshalStruct(e *Encoder, rv reflect.Value) error {
	rt := rv.Type()
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
		if err := marshalFieldInline(e, rv.Field(i), tag, &deferred); err != nil {
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
func marshalFieldInline(e *Encoder, fv reflect.Value, tag fieldTag, deferred *[]func() error) error {
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
			// Reference pointers carry no referent id at the top level; the referent
			// is marshalled inline.
			if isNilReferent(fv) {
				return fmt.Errorf("ndr: nil [ref] pointer")
			}
			return marshalReferentBody(e, fv, tag)
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
		return marshalStruct(e, fv)
	case reflect.Slice:
		if fv.Type().Elem().Kind() == reflect.Uint8 {
			e.writeConformantBytes(fv.Bytes())
			return nil
		}
		return fmt.Errorf("ndr: unsupported slice element %s", fv.Type().Elem().Kind())
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

func unmarshalStruct(d *Decoder, rv reflect.Value) error {
	rt := rv.Type()
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
		if err := unmarshalFieldInline(d, rv.Field(i), tag, &deferred); err != nil {
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

func unmarshalFieldInline(d *Decoder, fv reflect.Value, tag fieldTag, deferred *[]func() error) error {
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
			return unmarshalReferentBody(d, fv, tag)
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
		return unmarshalStruct(d, fv)
	case reflect.Slice:
		if fv.Type().Elem().Kind() == reflect.Uint8 {
			b, err := d.readConformantBytes()
			if err != nil {
				return err
			}
			fv.SetBytes(b)
			return nil
		}
		return fmt.Errorf("ndr: unsupported slice element %s", fv.Type().Elem().Kind())
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

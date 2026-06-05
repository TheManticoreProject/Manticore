package ndr

import (
	"reflect"
	"strconv"
	"strings"
)

// Windows-style NDR type aliases, mirroring the names used in [MS-DTYP] and impacket
// so declarations read like the IDL. References:
//   - [MS-DTYP] 2.2 Common Data Types:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/cca27429-5689-4a16-b2b4-9325d93e4ba2
type (
	BYTE    = uint8
	WORD    = uint16
	DWORD   = uint32
	DWORD64 = uint64
	LONG    = int32

	// BOOL is the Windows BOOL data type, a 4-octet integer ([MS-DTYP] 2.2.3). It is
	// distinct from the 1-octet NDR boolean: for the latter, use Go bool (or BOOLEAN).
	BOOL = int32
	// BOOLEAN is the 1-octet NDR boolean ([C706] section 14.2.4).
	BOOLEAN = bool

	// WSTR is a wchar_t string ([string] wide). A field of this type marshals as a
	// conformant+varying UTF-16LE array without needing an explicit "wstr" tag.
	//
	// Pointer semantics are explicit and controlled by the tag: a bare WSTR/STR/string
	// field is encoded inline with no referent id (matching a top-level [ref] string
	// parameter), while one tagged "unique" or "ptr" is encoded as a referent id with
	// its body deferred (matching an embedded [unique,string] wchar_t*). Embedded
	// strings in MS-RPC structures are almost always [unique]; tag them accordingly.
	WSTR string
	// STR is a char string ([string]). A field of this type marshals as a
	// conformant+varying ASCII array. See WSTR for pointer semantics.
	STR string
)

// wstrType and strType are cached reflect.Types for the named string aliases.
var (
	wstrType = reflect.TypeOf(WSTR(""))
	strType  = reflect.TypeOf(STR(""))
)

// Call is implemented by a request structure to provide its operation number. Its
// exported fields are the [in] parameters, marshalled in declaration order.
type Call interface {
	Opnum() uint16
}

// Invoker executes an RPC call: it marshals in (an ndr.Call) to a request stub,
// transmits it on opnum in.Opnum(), and unmarshals the response stub into out. It is the
// transport-version-neutral surface that interface method stubs depend on, so the stubs
// stay independent of any concrete client or wire-protocol version. Both the
// connection-oriented (network/dcerpc/v5/client) and, in future, connectionless
// (network/dcerpc/v4) DCE/RPC clients satisfy it.
type Invoker interface {
	Invoke(in Call, out any) error
}

// Request marshals an RPC call's [in] parameters into a stub buffer suitable for
// network/dcerpc/client.Client.Call(call.Opnum(), stub).
func Request(call Call) ([]byte, error) {
	return Marshal(call)
}

// Response unmarshals an RPC response stub into out (a pointer to the [out] parameter
// structure).
func Response(stub []byte, out any) error {
	return Unmarshal(stub, out)
}

// ptrKind is the NDR pointer attribute of a field.
type ptrKind int

const (
	ptrNone   ptrKind = iota // not a pointer
	ptrUnique                // [unique]
	ptrRef                   // [ref]
	ptrFull                  // [ptr] (full pointer)
)

// fieldTag is the parsed `ndr:"..."` struct tag.
type fieldTag struct {
	skip         bool
	ptr          ptrKind
	wide         bool    // [string] wide (UTF-16)
	ascii        bool    // [string] ASCII
	conformant   bool    // conformant array (size_is)
	varying      bool    // conformant-varying array (offset + actual_count framing)
	sizeIs       string  // sibling field naming the maximum element count
	lengthIs     string  // sibling field naming the actual element count
	sizeConst    uint32  // literal maximum_count (size_is(<N>)); used when sizeConstSet
	sizeConstSet bool    // size_is was a numeric literal rather than a sibling field name
	align        int     // explicit alignment override (0 = default)
	retval       bool    // RPC return value: encoded after the struct's deferred referents
	elemPtr      ptrKind // pointer attribute of array elements (`elem=ref|unique|ptr`)
	pipe         bool    // NDR pipe: a chunked stream ([C706] 14.7), not a normal array

	// Union (discriminated by an inline switch value, [C706] section 14.3.8) tags.
	isSwitch  bool  // the union discriminant field (`switch`)
	hasCase   bool  // this field is a union arm selected by `case=<value>`
	caseVal   int64 // the discriminant value that selects this arm
	isDefault bool  // the union's default arm (`default`)
}

// parseTag parses an `ndr:"..."` tag value.
func parseTag(raw string) fieldTag {
	var t fieldTag
	if raw == "-" {
		t.skip = true
		return t
	}
	for _, opt := range strings.Split(raw, ",") {
		opt = strings.TrimSpace(opt)
		switch {
		case opt == "":
		case opt == "unique":
			t.ptr = ptrUnique
		case opt == "ref":
			t.ptr = ptrRef
		case opt == "ptr", opt == "full":
			t.ptr = ptrFull
		case opt == "wstr":
			t.wide = true
		case opt == "str":
			t.ascii = true
		case opt == "pipe":
			t.pipe = true
		case opt == "conformant":
			t.conformant = true
		case opt == "varying":
			t.varying = true
		case opt == "retval":
			t.retval = true
		case strings.HasPrefix(opt, "size_is="):
			t.conformant = true
			v := strings.TrimPrefix(opt, "size_is=")
			// size_is(<constant>) — a literal maximum_count, e.g. [size_is(1000)] in
			// MS-SAMR. Distinguished from a sibling field name by parsing as a number.
			if n, err := strconv.ParseUint(v, 10, 32); err == nil {
				t.sizeConst = uint32(n)
				t.sizeConstSet = true
			} else {
				t.sizeIs = v
			}
		case strings.HasPrefix(opt, "length_is="):
			t.conformant = true
			t.varying = true
			t.lengthIs = strings.TrimPrefix(opt, "length_is=")
		case opt == "switch":
			t.isSwitch = true
		case opt == "default":
			t.isDefault = true
		case strings.HasPrefix(opt, "case="):
			if v, err := strconv.ParseInt(strings.TrimPrefix(opt, "case="), 0, 64); err == nil {
				t.hasCase = true
				t.caseVal = v
			}
		case opt == "elem=ref":
			t.elemPtr = ptrRef
		case opt == "elem=unique":
			t.elemPtr = ptrUnique
		case opt == "elem=ptr", opt == "elem=full":
			t.elemPtr = ptrFull
		case strings.HasPrefix(opt, "align="):
			switch strings.TrimPrefix(opt, "align=") {
			case "1":
				t.align = 1
			case "2":
				t.align = 2
			case "4":
				t.align = 4
			case "8":
				t.align = 8
			}
		// Scalar hints (dword/word/byte/hyper/bool) are accepted for documentation
		// but the wire type is inferred from the Go kind, so they need no handling.
		default:
		}
	}
	return t
}

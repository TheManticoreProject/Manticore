package ms_rrp

import (
	"encoding/binary"
	"strings"
	"unicode/utf16"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// RegistryValue is a typed registry value: its REG_* type and its raw little-endian data
// bytes, exactly as carried on the wire. The decode helpers (String, Uint32, ...) project
// the bytes per Type; the constructor helpers (StringValue, DwordValue, ...) build the
// bytes for SetValue.
type RegistryValue struct {
	Type uint32
	Data []byte
}

// String decodes REG_SZ / REG_EXPAND_SZ / REG_LINK as a UTF-16LE string (trailing NULs
// stripped). For other types it returns the empty string; use the type-specific helpers.
func (v RegistryValue) String() string {
	switch v.Type {
	case winreg.RegSz, winreg.RegExpandSz, winreg.RegLink:
		return decodeUTF16(v.Data)
	default:
		return ""
	}
}

// Uint32 decodes a REG_DWORD (little-endian). ok is false if the data is too short.
func (v RegistryValue) Uint32() (uint32, bool) {
	if len(v.Data) < 4 {
		return 0, false
	}
	return binary.LittleEndian.Uint32(v.Data), true
}

// Uint64 decodes a REG_QWORD (little-endian). ok is false if the data is too short.
func (v RegistryValue) Uint64() (uint64, bool) {
	if len(v.Data) < 8 {
		return 0, false
	}
	return binary.LittleEndian.Uint64(v.Data), true
}

// MultiString decodes a REG_MULTI_SZ: a sequence of UTF-16LE strings, each NUL-terminated,
// the whole terminated by an empty string. Empty trailing entries are dropped.
func (v RegistryValue) MultiString() []string {
	if v.Type != winreg.RegMultiSz {
		return nil
	}
	whole := decodeUTF16Raw(v.Data)
	parts := strings.Split(whole, "\x00")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

// StringValue builds a REG_SZ value from a Go string.
func StringValue(s string) RegistryValue {
	return RegistryValue{Type: winreg.RegSz, Data: encodeUTF16(s)}
}

// ExpandStringValue builds a REG_EXPAND_SZ value from a Go string.
func ExpandStringValue(s string) RegistryValue {
	return RegistryValue{Type: winreg.RegExpandSz, Data: encodeUTF16(s)}
}

// DwordValue builds a REG_DWORD value (little-endian).
func DwordValue(v uint32) RegistryValue {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return RegistryValue{Type: winreg.RegDword, Data: b}
}

// QwordValue builds a REG_QWORD value (little-endian).
func QwordValue(v uint64) RegistryValue {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	return RegistryValue{Type: winreg.RegQword, Data: b}
}

// BinaryValue builds a REG_BINARY value from raw bytes.
func BinaryValue(b []byte) RegistryValue {
	return RegistryValue{Type: winreg.RegBinary, Data: append([]byte(nil), b...)}
}

// MultiStringValue builds a REG_MULTI_SZ value: each string NUL-terminated, the block
// terminated by a final empty string.
func MultiStringValue(items []string) RegistryValue {
	return RegistryValue{Type: winreg.RegMultiSz, Data: encodeUTF16Multi(items)}
}

// ValueEntry pairs a value's name with its typed data, as returned by EnumValues.
type ValueEntry struct {
	Name  string
	Value RegistryValue
}

// BaseRegQueryValue calls BaseRegQueryValue (opnum 17). It mirrors the interface function
// exactly (minus the invoker); for ergonomic reads use QueryValueByPath / queryValue.
func (r *RemoteRegistry) BaseRegQueryValue(hKey structures.RPC_HKEY, lpValueName structures.RRP_UNICODE_STRING, lpType *ndr.DWORD, lpData []uint8, lpcbData *ndr.DWORD, lpcbLen *ndr.DWORD) (*ndr.DWORD, []uint8, *ndr.DWORD, *ndr.DWORD, error) {
	if err := r.ensure(); err != nil {
		return nil, nil, nil, nil, err
	}
	return functions.BaseRegQueryValue(r.rpc, hKey, lpValueName, lpType, lpData, lpcbData, lpcbLen)
}

// BaseRegSetValue calls BaseRegSetValue (opnum 22): writes a value's type and data.
func (r *RemoteRegistry) BaseRegSetValue(hKey structures.RPC_HKEY, lpValueName structures.RRP_UNICODE_STRING, dwType ndr.DWORD, lpData []uint8, cbData ndr.DWORD) error {
	if err := r.ensure(); err != nil {
		return err
	}
	return functions.BaseRegSetValue(r.rpc, hKey, lpValueName, dwType, lpData, cbData)
}

// BaseRegDeleteValue calls BaseRegDeleteValue (opnum 8): removes a named value.
func (r *RemoteRegistry) BaseRegDeleteValue(hKey structures.RPC_HKEY, lpValueName structures.RRP_UNICODE_STRING) error {
	if err := r.ensure(); err != nil {
		return err
	}
	return functions.BaseRegDeleteValue(r.rpc, hKey, lpValueName)
}

// BaseRegEnumValue calls BaseRegEnumValue (opnum 10): returns the value at dwIndex.
func (r *RemoteRegistry) BaseRegEnumValue(hKey structures.RPC_HKEY, dwIndex ndr.DWORD, lpValueNameIn structures.RRP_UNICODE_STRING, lpType *ndr.DWORD, lpData []uint8, lpcbData *ndr.DWORD, lpcbLen *ndr.DWORD) (dtyp.RPC_UNICODE_STRING, *ndr.DWORD, []uint8, *ndr.DWORD, *ndr.DWORD, error) {
	if err := r.ensure(); err != nil {
		return dtyp.RPC_UNICODE_STRING{}, nil, nil, nil, nil, err
	}
	return functions.BaseRegEnumValue(r.rpc, hKey, dwIndex, lpValueNameIn, lpType, lpData, lpcbData, lpcbLen)
}

// BaseRegQueryMultipleValues calls BaseRegQueryMultipleValues (opnum 29).
func (r *RemoteRegistry) BaseRegQueryMultipleValues(hKey structures.RPC_HKEY, val_listIn []structures.RVALENT, num_vals ndr.DWORD, lpvalueBuf []byte, ldwTotsize ndr.DWORD) ([]structures.RVALENT, []byte, ndr.DWORD, error) {
	if err := r.ensure(); err != nil {
		return nil, nil, 0, err
	}
	return functions.BaseRegQueryMultipleValues(r.rpc, hKey, val_listIn, num_vals, lpvalueBuf, ldwTotsize)
}

// BaseRegQueryMultipleValues2 calls BaseRegQueryMultipleValues2 (opnum 34).
func (r *RemoteRegistry) BaseRegQueryMultipleValues2(hKey structures.RPC_HKEY, val_listIn []structures.RVALENT, num_vals ndr.DWORD, lpvalueBuf []byte, ldwTotsize ndr.DWORD) ([]structures.RVALENT, []byte, ndr.DWORD, error) {
	if err := r.ensure(); err != nil {
		return nil, nil, 0, err
	}
	return functions.BaseRegQueryMultipleValues2(r.rpc, hKey, val_listIn, num_vals, lpvalueBuf, ldwTotsize)
}

// --- UTF-16LE helpers ---

// encodeUTF16 encodes s as UTF-16LE with a terminating NUL (the form REG_SZ uses).
func encodeUTF16(s string) []byte {
	u := utf16.Encode([]rune(s))
	b := make([]byte, len(u)*2+2)
	for i, c := range u {
		binary.LittleEndian.PutUint16(b[i*2:], c)
	}
	return b
}

// encodeUTF16Multi encodes a REG_MULTI_SZ block: each item NUL-terminated, a final empty
// string closing the block.
func encodeUTF16Multi(items []string) []byte {
	var b []byte
	for _, s := range items {
		b = append(b, encodeUTF16(s)...)
	}
	b = append(b, 0x00, 0x00) // final empty string terminator
	return b
}

// decodeUTF16 decodes UTF-16LE bytes and strips a single trailing NUL terminator.
func decodeUTF16(b []byte) string {
	return strings.TrimRight(decodeUTF16Raw(b), "\x00")
}

// decodeUTF16Raw decodes UTF-16LE bytes without stripping NULs (an odd trailing byte is
// ignored).
func decodeUTF16Raw(b []byte) string {
	n := len(b) / 2
	u := make([]uint16, n)
	for i := 0; i < n; i++ {
		u[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	return string(utf16.Decode(u))
}

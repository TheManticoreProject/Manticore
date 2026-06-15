package ms_rrp

import (
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// This file holds the ergonomic, reg.exe-style helpers layered on top of the one-to-one
// interface methods. They accept Go string paths ("HKLM\\Software\\...") and value names,
// open and close the necessary handles, and project results into Go types. The method
// names here are deliberately distinct from the interface method names.

// maximumAllowed is MAXIMUM_ALLOWED ([MS-DTYP] 2.4.3): asks the server to grant the most
// access the caller is permitted. Used when opening a transient root handle whose only job
// is to host a create/delete of a subkey beneath it.
const maximumAllowed ndr.DWORD = 0x02000000

// regName builds a NUL-terminated RRP_UNICODE_STRING for a key/value name. [MS-RRP] 3.1.1
// counts the terminating NUL in the length, so it is appended explicitly.
func regName(s string) structures.RRP_UNICODE_STRING {
	return dtyp.NewUnicodeString(s + "\x00")
}

// isStatus reports whether err carries the given winreg Win32 status code, matched by the
// mnemonic the interface stubs embed in their error text (e.g. ERROR_NO_MORE_ITEMS). This
// lets enumeration/query loops treat the documented terminal codes as sentinels rather
// than hard failures.
func isStatus(err error, code uint32) bool {
	return err != nil && strings.Contains(err.Error(), winreg.StatusString(code))
}

// splitRegistryPath splits "HKLM\\Software\\Foo" into the root mnemonic ("HKLM") and the
// remaining subkey path ("Software\\Foo"). Leading/trailing backslashes and spaces are
// trimmed; a path with no subkey returns an empty subkey.
func splitRegistryPath(path string) (root, subkey string) {
	path = strings.Trim(strings.TrimSpace(path), `\`)
	if i := strings.IndexByte(path, '\\'); i >= 0 {
		return path[:i], path[i+1:]
	}
	return path, ""
}

// openRoot opens the predefined root key named by root (HKLM, HKEY_LOCAL_MACHINE, HKCU,
// HKCR, HKU, HKCC, HKPD and their long forms), returning its handle.
func (r *RemoteRegistry) openRoot(root string, samDesired ndr.DWORD) (Handle, error) {
	switch strings.ToUpper(strings.TrimSpace(root)) {
	case "HKLM", "HKEY_LOCAL_MACHINE":
		return r.OpenLocalMachine(nil, samDesired)
	case "HKCU", "HKEY_CURRENT_USER":
		return r.OpenCurrentUser(nil, samDesired)
	case "HKCR", "HKEY_CLASSES_ROOT":
		return r.OpenClassesRoot(nil, samDesired)
	case "HKU", "HKEY_USERS":
		return r.OpenUsers(nil, samDesired)
	case "HKCC", "HKEY_CURRENT_CONFIG":
		return r.OpenCurrentConfig(nil, samDesired)
	case "HKPD", "HKEY_PERFORMANCE_DATA":
		return r.OpenPerformanceData(nil, samDesired)
	default:
		return Handle{}, fmt.Errorf("ms_rrp: unknown registry root %q", root)
	}
}

// OpenKeyByPath opens the key at a full path such as "HKLM\\SOFTWARE\\Microsoft", with the
// requested access. The caller owns the returned handle and must BaseRegCloseKey it. The
// transient root handle is closed before returning.
func (r *RemoteRegistry) OpenKeyByPath(keyPath string, samDesired ndr.DWORD) (Handle, error) {
	root, sub := splitRegistryPath(keyPath)
	h, err := r.openRoot(root, samDesired)
	if err != nil {
		return Handle{}, err
	}
	if sub == "" {
		return h, nil
	}
	sk, err := r.BaseRegOpenKey(h, regName(sub), 0, samDesired)
	_, _ = r.BaseRegCloseKey(h) // the subkey handle is independent of the root handle
	if err != nil {
		return Handle{}, err
	}
	return sk, nil
}

// queryValue reads a single value under an open key, negotiating the data buffer size: it
// grows and retries on ERROR_MORE_DATA.
func (r *RemoteRegistry) queryValue(h Handle, valueName string) (RegistryValue, error) {
	bufLen := uint32(256)
	for attempts := 0; attempts < 5; attempts++ {
		buf := make([]byte, bufLen)
		typ := ndr.DWORD(0)
		cb := ndr.DWORD(bufLen)
		ln := ndr.DWORD(bufLen)
		rTyp, rData, rcb, _, err := r.BaseRegQueryValue(h, regName(valueName), &typ, buf, &cb, &ln)
		if err != nil {
			if isStatus(err, winreg.ErrorMoreData) && rcb != nil && uint32(*rcb) > bufLen {
				bufLen = uint32(*rcb)
				continue
			}
			return RegistryValue{}, err
		}
		n := bufLen
		if rcb != nil {
			n = uint32(*rcb)
		}
		if int(n) > len(rData) {
			n = uint32(len(rData))
		}
		out := RegistryValue{Data: append([]byte(nil), rData[:n]...)}
		if rTyp != nil {
			out.Type = uint32(*rTyp)
		}
		return out, nil
	}
	return RegistryValue{}, fmt.Errorf("ms_rrp: value %q under the key kept requesting a larger buffer", valueName)
}

// QueryValueByPath opens keyPath for read, reads valueName, and closes the key.
func (r *RemoteRegistry) QueryValueByPath(keyPath, valueName string) (RegistryValue, error) {
	h, err := r.OpenKeyByPath(keyPath, winreg.KeyQueryValue)
	if err != nil {
		return RegistryValue{}, err
	}
	defer r.BaseRegCloseKey(h)
	return r.queryValue(h, valueName)
}

// SetValueByPath opens keyPath for write, writes value under valueName, and closes the key.
func (r *RemoteRegistry) SetValueByPath(keyPath, valueName string, value RegistryValue) error {
	h, err := r.OpenKeyByPath(keyPath, winreg.KeySetValue)
	if err != nil {
		return err
	}
	defer r.BaseRegCloseKey(h)
	return r.BaseRegSetValue(h, regName(valueName), ndr.DWORD(value.Type), value.Data, ndr.DWORD(len(value.Data)))
}

// DeleteValueByPath opens keyPath for write, removes valueName, and closes the key.
func (r *RemoteRegistry) DeleteValueByPath(keyPath, valueName string) error {
	h, err := r.OpenKeyByPath(keyPath, winreg.KeySetValue)
	if err != nil {
		return err
	}
	defer r.BaseRegCloseKey(h)
	return r.BaseRegDeleteValue(h, regName(valueName))
}

// EnumKeys returns the names of all immediate subkeys of an open key, looping
// BaseRegEnumKey until ERROR_NO_MORE_ITEMS. Subkey names are capped at 255 UTF-16 code
// units by the registry, but the name/class buffers grow and retry the same index on
// ERROR_MORE_DATA to stay robust if a server reports a longer name.
func (r *RemoteRegistry) EnumKeys(h Handle) ([]string, error) {
	var names []string
	// bufLen counts UTF-16 code units; MaximumLength is that count in bytes (×2). The cap
	// keeps MaximumLength within its uint16 field while far exceeding the 255-char limit.
	bufLen := uint32(256)
	const maxBufLen = uint32(16 * 1024)
	for i := uint32(0); ; {
		nameIn := structures.RRP_UNICODE_STRING{MaximumLength: uint16(bufLen * 2), Buffer: make([]uint16, bufLen)}
		classIn := structures.RRP_UNICODE_STRING{MaximumLength: uint16(bufLen * 2), Buffer: make([]uint16, bufLen)}
		nameOut, _, _, err := r.BaseRegEnumKey(h, ndr.DWORD(i), nameIn, &classIn, nil)
		if err != nil {
			if isStatus(err, winreg.ErrorNoMoreItems) {
				break
			}
			if isStatus(err, winreg.ErrorMoreData) && bufLen < maxBufLen {
				// The name (or class) did not fit: grow the buffers and retry the same index.
				bufLen *= 2
				continue
			}
			return names, err
		}
		// The server counts the terminating NUL in the returned name's Length, so strip it
		// to yield a clean Go string usable as a subkey name / path component.
		names = append(names, strings.TrimRight(nameOut.String(), "\x00"))
		i++
	}
	return names, nil
}

// EnumKeysByPath opens keyPath, enumerates its subkeys, and closes the key.
func (r *RemoteRegistry) EnumKeysByPath(keyPath string) ([]string, error) {
	h, err := r.OpenKeyByPath(keyPath, winreg.KeyEnumerateSubKeys)
	if err != nil {
		return nil, err
	}
	defer r.BaseRegCloseKey(h)
	return r.EnumKeys(h)
}

// EnumValues returns all values of an open key, looping BaseRegEnumValue until
// ERROR_NO_MORE_ITEMS. The data buffer is negotiated per value: it grows and retries the
// same index on ERROR_MORE_DATA (e.g. values whose data exceeds the initial buffer).
func (r *RemoteRegistry) EnumValues(h Handle) ([]ValueEntry, error) {
	var entries []ValueEntry
	dataLen := uint32(1024)
	for i := uint32(0); ; {
		// A registry value name fits in 1023 UTF-16 code units here; only the data buffer
		// needs to grow, so the name buffer is fixed and the data buffer is negotiated.
		nameIn := structures.RRP_UNICODE_STRING{MaximumLength: 2048, Buffer: make([]uint16, 1024)}
		typ := ndr.DWORD(0)
		buf := make([]byte, dataLen)
		cb := ndr.DWORD(dataLen)
		ln := ndr.DWORD(dataLen)
		nameOut, rTyp, rData, rcb, _, err := r.BaseRegEnumValue(h, ndr.DWORD(i), nameIn, &typ, buf, &cb, &ln)
		if err != nil {
			if isStatus(err, winreg.ErrorNoMoreItems) {
				break
			}
			if isStatus(err, winreg.ErrorMoreData) && rcb != nil && uint32(*rcb) > dataLen {
				// The value's data is larger than the current buffer: grow it and retry the
				// same index without advancing.
				dataLen = uint32(*rcb)
				continue
			}
			return entries, err
		}
		n := len(rData)
		if rcb != nil && int(*rcb) < n {
			n = int(*rcb)
		}
		val := RegistryValue{Data: append([]byte(nil), rData[:n]...)}
		if rTyp != nil {
			val.Type = uint32(*rTyp)
		}
		// The server counts the terminating NUL in the returned name's Length; strip it so
		// the value name is a clean Go string.
		entries = append(entries, ValueEntry{Name: strings.TrimRight(nameOut.String(), "\x00"), Value: val})
		i++
	}
	return entries, nil
}

// EnumValuesByPath opens keyPath, enumerates its values, and closes the key.
func (r *RemoteRegistry) EnumValuesByPath(keyPath string) ([]ValueEntry, error) {
	h, err := r.OpenKeyByPath(keyPath, winreg.KeyQueryValue)
	if err != nil {
		return nil, err
	}
	defer r.BaseRegCloseKey(h)
	return r.EnumValues(h)
}

// CreateKeyByPath creates (or opens, if it already exists) the key at keyPath, returning
// the new handle and the disposition (RegCreatedNewKey / RegOpenedExistingKey). The caller
// must BaseRegCloseKey the returned handle.
func (r *RemoteRegistry) CreateKeyByPath(keyPath string, samDesired ndr.DWORD) (Handle, ndr.DWORD, error) {
	root, sub := splitRegistryPath(keyPath)
	if sub == "" {
		return Handle{}, 0, fmt.Errorf("ms_rrp: cannot create a root key %q", root)
	}
	rootHandle, err := r.openRoot(root, maximumAllowed)
	if err != nil {
		return Handle{}, 0, err
	}
	defer r.BaseRegCloseKey(rootHandle)
	disp := ndr.DWORD(0)
	sk, rdisp, err := r.BaseRegCreateKey(rootHandle, regName(sub), structures.RRP_UNICODE_STRING{}, ndr.DWORD(winreg.RegOptionNonVolatile), samDesired, nil, &disp)
	if err != nil {
		return Handle{}, 0, err
	}
	d := disp
	if rdisp != nil {
		d = *rdisp
	}
	return sk, d, nil
}

// DeleteKeyByPath deletes the key at keyPath. The key must have no subkeys (per
// BaseRegDeleteKey). Root keys cannot be deleted.
func (r *RemoteRegistry) DeleteKeyByPath(keyPath string) error {
	root, sub := splitRegistryPath(keyPath)
	if sub == "" {
		return fmt.Errorf("ms_rrp: refusing to delete root key %q", root)
	}
	rootHandle, err := r.openRoot(root, maximumAllowed)
	if err != nil {
		return err
	}
	defer r.BaseRegCloseKey(rootHandle)
	return r.BaseRegDeleteKey(rootHandle, regName(sub))
}

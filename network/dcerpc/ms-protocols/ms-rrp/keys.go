package ms_rrp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// The methods in this file mirror the winreg interface functions one-to-one: same name,
// same parameters and returns, minus the leading rpc invoker (the bound association is
// taken from the RemoteRegistry). They are the faithful, low-level surface; the
// ergonomic *ByPath helpers in paths.go are built on top of them.

// OpenClassesRoot calls OpenClassesRoot (opnum 0): opens HKEY_CLASSES_ROOT.
func (r *RemoteRegistry) OpenClassesRoot(serverName *ndr.WSTR, samDesired ndr.DWORD) (msrrp.PRPC_HKEY, error) {
	if err := r.ensure(); err != nil {
		return Handle{}, err
	}
	return functions.OpenClassesRoot(r.rpc, serverName, samDesired)
}

// OpenCurrentUser calls OpenCurrentUser (opnum 1): opens HKEY_CURRENT_USER.
func (r *RemoteRegistry) OpenCurrentUser(serverName *ndr.WSTR, samDesired ndr.DWORD) (msrrp.PRPC_HKEY, error) {
	if err := r.ensure(); err != nil {
		return Handle{}, err
	}
	return functions.OpenCurrentUser(r.rpc, serverName, samDesired)
}

// OpenLocalMachine calls OpenLocalMachine (opnum 2): opens HKEY_LOCAL_MACHINE.
func (r *RemoteRegistry) OpenLocalMachine(serverName *ndr.WSTR, samDesired ndr.DWORD) (msrrp.PRPC_HKEY, error) {
	if err := r.ensure(); err != nil {
		return Handle{}, err
	}
	return functions.OpenLocalMachine(r.rpc, serverName, samDesired)
}

// OpenPerformanceData calls OpenPerformanceData (opnum 3): opens HKEY_PERFORMANCE_DATA.
func (r *RemoteRegistry) OpenPerformanceData(serverName *ndr.WSTR, samDesired ndr.DWORD) (msrrp.PRPC_HKEY, error) {
	if err := r.ensure(); err != nil {
		return Handle{}, err
	}
	return functions.OpenPerformanceData(r.rpc, serverName, samDesired)
}

// OpenUsers calls OpenUsers (opnum 4): opens HKEY_USERS.
func (r *RemoteRegistry) OpenUsers(serverName *ndr.WSTR, samDesired ndr.DWORD) (msrrp.PRPC_HKEY, error) {
	if err := r.ensure(); err != nil {
		return Handle{}, err
	}
	return functions.OpenUsers(r.rpc, serverName, samDesired)
}

// OpenCurrentConfig calls OpenCurrentConfig (opnum 27): opens HKEY_CURRENT_CONFIG.
func (r *RemoteRegistry) OpenCurrentConfig(serverName *ndr.WSTR, samDesired ndr.DWORD) (msrrp.PRPC_HKEY, error) {
	if err := r.ensure(); err != nil {
		return Handle{}, err
	}
	return functions.OpenCurrentConfig(r.rpc, serverName, samDesired)
}

// OpenPerformanceText calls OpenPerformanceText (opnum 32).
func (r *RemoteRegistry) OpenPerformanceText(serverName *ndr.WSTR, samDesired ndr.DWORD) (msrrp.PRPC_HKEY, error) {
	if err := r.ensure(); err != nil {
		return Handle{}, err
	}
	return functions.OpenPerformanceText(r.rpc, serverName, samDesired)
}

// OpenPerformanceNlsText calls OpenPerformanceNlsText (opnum 33).
func (r *RemoteRegistry) OpenPerformanceNlsText(serverName *ndr.WSTR, samDesired ndr.DWORD) (msrrp.PRPC_HKEY, error) {
	if err := r.ensure(); err != nil {
		return Handle{}, err
	}
	return functions.OpenPerformanceNlsText(r.rpc, serverName, samDesired)
}

// BaseRegCloseKey calls BaseRegCloseKey (opnum 5): releases an open key handle.
func (r *RemoteRegistry) BaseRegCloseKey(hKey msrrp.PRPC_HKEY) (msrrp.PRPC_HKEY, error) {
	if err := r.ensure(); err != nil {
		return Handle{}, err
	}
	return functions.BaseRegCloseKey(r.rpc, hKey)
}

// BaseRegCreateKey calls BaseRegCreateKey (opnum 6): creates or opens a subkey.
func (r *RemoteRegistry) BaseRegCreateKey(hKey msrrp.RPC_HKEY, lpSubKey msrrp.RRP_UNICODE_STRING, lpClass msrrp.RRP_UNICODE_STRING, dwOptions ndr.DWORD, samDesired ndr.DWORD, lpSecurityAttributes *msrrp.RPC_SECURITY_ATTRIBUTES, lpdwDisposition *ndr.DWORD) (msrrp.PRPC_HKEY, *ndr.DWORD, error) {
	if err := r.ensure(); err != nil {
		return Handle{}, nil, err
	}
	return functions.BaseRegCreateKey(r.rpc, hKey, lpSubKey, lpClass, dwOptions, samDesired, lpSecurityAttributes, lpdwDisposition)
}

// BaseRegDeleteKey calls BaseRegDeleteKey (opnum 7): deletes a subkey with no subkeys.
func (r *RemoteRegistry) BaseRegDeleteKey(hKey msrrp.RPC_HKEY, lpSubKey msrrp.RRP_UNICODE_STRING) error {
	if err := r.ensure(); err != nil {
		return err
	}
	return functions.BaseRegDeleteKey(r.rpc, hKey, lpSubKey)
}

// BaseRegEnumKey calls BaseRegEnumKey (opnum 9): returns the subkey at dwIndex.
func (r *RemoteRegistry) BaseRegEnumKey(hKey msrrp.RPC_HKEY, dwIndex ndr.DWORD, lpNameIn msrrp.RRP_UNICODE_STRING, lpClassIn *msrrp.RRP_UNICODE_STRING, lpftLastWriteTime *msdtyp.FILETIME) (msrrp.RRP_UNICODE_STRING, *msdtyp.RPC_UNICODE_STRING, *msdtyp.FILETIME, error) {
	if err := r.ensure(); err != nil {
		return msrrp.RRP_UNICODE_STRING{}, nil, nil, err
	}
	return functions.BaseRegEnumKey(r.rpc, hKey, dwIndex, lpNameIn, lpClassIn, lpftLastWriteTime)
}

// BaseRegFlushKey calls BaseRegFlushKey (opnum 11): writes a key's changes to disk.
func (r *RemoteRegistry) BaseRegFlushKey(hKey msrrp.RPC_HKEY) error {
	if err := r.ensure(); err != nil {
		return err
	}
	return functions.BaseRegFlushKey(r.rpc, hKey)
}

// BaseRegLoadKey calls BaseRegLoadKey (opnum 13): loads a hive file under a subkey.
func (r *RemoteRegistry) BaseRegLoadKey(hKey msrrp.RPC_HKEY, lpSubKey msrrp.RRP_UNICODE_STRING, lpFile msrrp.RRP_UNICODE_STRING) error {
	if err := r.ensure(); err != nil {
		return err
	}
	return functions.BaseRegLoadKey(r.rpc, hKey, lpSubKey, lpFile)
}

// BaseRegOpenKey calls BaseRegOpenKey (opnum 15): opens an existing subkey.
func (r *RemoteRegistry) BaseRegOpenKey(hKey msrrp.RPC_HKEY, lpSubKey msrrp.RRP_UNICODE_STRING, dwOptions ndr.DWORD, samDesired ndr.DWORD) (msrrp.PRPC_HKEY, error) {
	if err := r.ensure(); err != nil {
		return Handle{}, err
	}
	return functions.BaseRegOpenKey(r.rpc, hKey, lpSubKey, dwOptions, samDesired)
}

// BaseRegQueryInfoKey calls BaseRegQueryInfoKey (opnum 16): returns subkey/value counts,
// maximum lengths, and the last-write time of a key.
func (r *RemoteRegistry) BaseRegQueryInfoKey(hKey msrrp.RPC_HKEY, lpClassIn msrrp.RRP_UNICODE_STRING) (msdtyp.RPC_UNICODE_STRING, ndr.DWORD, ndr.DWORD, ndr.DWORD, ndr.DWORD, ndr.DWORD, ndr.DWORD, ndr.DWORD, msdtyp.FILETIME, error) {
	if err := r.ensure(); err != nil {
		return msdtyp.RPC_UNICODE_STRING{}, 0, 0, 0, 0, 0, 0, 0, msdtyp.FILETIME{}, err
	}
	return functions.BaseRegQueryInfoKey(r.rpc, hKey, lpClassIn)
}

// BaseRegReplaceKey calls BaseRegReplaceKey (opnum 18).
func (r *RemoteRegistry) BaseRegReplaceKey(hKey msrrp.RPC_HKEY, lpSubKey msrrp.RRP_UNICODE_STRING, lpNewFile msrrp.RRP_UNICODE_STRING, lpOldFile msrrp.RRP_UNICODE_STRING) error {
	if err := r.ensure(); err != nil {
		return err
	}
	return functions.BaseRegReplaceKey(r.rpc, hKey, lpSubKey, lpNewFile, lpOldFile)
}

// BaseRegRestoreKey calls BaseRegRestoreKey (opnum 19).
func (r *RemoteRegistry) BaseRegRestoreKey(hKey msrrp.RPC_HKEY, lpFile msrrp.RRP_UNICODE_STRING, flags ndr.DWORD) error {
	if err := r.ensure(); err != nil {
		return err
	}
	return functions.BaseRegRestoreKey(r.rpc, hKey, lpFile, flags)
}

// BaseRegSaveKey calls BaseRegSaveKey (opnum 20): saves a key and its subtree to a file.
func (r *RemoteRegistry) BaseRegSaveKey(hKey msrrp.RPC_HKEY, lpFile msrrp.RRP_UNICODE_STRING, pSecurityAttributes *msrrp.RPC_SECURITY_ATTRIBUTES) error {
	if err := r.ensure(); err != nil {
		return err
	}
	return functions.BaseRegSaveKey(r.rpc, hKey, lpFile, pSecurityAttributes)
}

// BaseRegUnLoadKey calls BaseRegUnLoadKey (opnum 23): unloads a previously loaded hive.
func (r *RemoteRegistry) BaseRegUnLoadKey(hKey msrrp.RPC_HKEY, lpSubKey msrrp.RRP_UNICODE_STRING) error {
	if err := r.ensure(); err != nil {
		return err
	}
	return functions.BaseRegUnLoadKey(r.rpc, hKey, lpSubKey)
}

// BaseRegGetVersion calls BaseRegGetVersion (opnum 26): returns the registry version.
func (r *RemoteRegistry) BaseRegGetVersion(hKey msrrp.RPC_HKEY) (ndr.DWORD, error) {
	if err := r.ensure(); err != nil {
		return 0, err
	}
	return functions.BaseRegGetVersion(r.rpc, hKey)
}

// BaseRegSaveKeyEx calls BaseRegSaveKeyEx (opnum 31): saves a key with format flags.
func (r *RemoteRegistry) BaseRegSaveKeyEx(hKey msrrp.RPC_HKEY, lpFile msrrp.RRP_UNICODE_STRING, pSecurityAttributes *msrrp.RPC_SECURITY_ATTRIBUTES, flags ndr.DWORD) error {
	if err := r.ensure(); err != nil {
		return err
	}
	return functions.BaseRegSaveKeyEx(r.rpc, hKey, lpFile, pSecurityAttributes, flags)
}

// BaseRegDeleteKeyEx calls BaseRegDeleteKeyEx (opnum 35): deletes a subkey honouring the
// KEY_WOW64_* view selected by accessMask.
func (r *RemoteRegistry) BaseRegDeleteKeyEx(hKey msrrp.RPC_HKEY, lpSubKey msrrp.RRP_UNICODE_STRING, accessMask ndr.DWORD, reserved ndr.DWORD) error {
	if err := r.ensure(); err != nil {
		return err
	}
	return functions.BaseRegDeleteKeyEx(r.rpc, hKey, lpSubKey, accessMask, reserved)
}

package ms_rrp

import (
	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// REGSAM access rights ([MS-RRP] 2.2.4), re-exported from the winreg interface and typed
// as ndr.DWORD so they can be passed directly as the samDesired argument of the Open* /
// BaseRegOpenKey / BaseRegCreateKey methods without conversion or importing the interface
// package.
const (
	KeyQueryValue       = ndr.DWORD(winreg.KeyQueryValue)
	KeySetValue         = ndr.DWORD(winreg.KeySetValue)
	KeyCreateSubKey     = ndr.DWORD(winreg.KeyCreateSubKey)
	KeyEnumerateSubKeys = ndr.DWORD(winreg.KeyEnumerateSubKeys)
	KeyNotify           = ndr.DWORD(winreg.KeyNotify)
	KeyCreateLink       = ndr.DWORD(winreg.KeyCreateLink)
	KeyWow6464Key       = ndr.DWORD(winreg.KeyWow64_64Key)
	KeyWow6432Key       = ndr.DWORD(winreg.KeyWow64_32Key)
	KeyRead             = ndr.DWORD(winreg.KeyRead)
	KeyWrite            = ndr.DWORD(winreg.KeyWrite)
	KeyExecute          = ndr.DWORD(winreg.KeyExecute)
	KeyAllAccess        = ndr.DWORD(winreg.KeyAllAccess)

	// MaximumAllowed is MAXIMUM_ALLOWED ([MS-DTYP] 2.4.3): grant the most access the caller
	// is permitted.
	MaximumAllowed = maximumAllowed
)

// dwOptions for BaseRegCreateKey ([MS-RRP] 3.1.5.7), re-exported and typed as ndr.DWORD.
const (
	RegOptionNonVolatile = ndr.DWORD(winreg.RegOptionNonVolatile)
	RegOptionVolatile    = ndr.DWORD(winreg.RegOptionVolatile)
)

// Registry value types (the RegistryValue.Type / dwType field) ([MS-RRP] 3.1.1.6),
// re-exported from the winreg interface as plain uint32 to match RegistryValue.Type.
const (
	RegNone     = winreg.RegNone
	RegSz       = winreg.RegSz
	RegExpandSz = winreg.RegExpandSz
	RegBinary   = winreg.RegBinary
	RegDword    = winreg.RegDword
	RegLink     = winreg.RegLink
	RegMultiSz  = winreg.RegMultiSz
	RegQword    = winreg.RegQword
)

// Key-creation disposition values returned by BaseRegCreateKey / CreateKeyByPath
// ([MS-RRP] 3.1.5.7).
const (
	RegCreatedNewKey     = winreg.RegCreatedNewKey
	RegOpenedExistingKey = winreg.RegOpenedExistingKey
)

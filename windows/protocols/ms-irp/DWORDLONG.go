package msirp

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// DWORDLONG is the [MS-DTYP] 2.2.6 unsigned 64-bit integer (unsigned __int64). It
// is imported from ms-dtyp.idl by [MS-IRP] rather than defined inline, so model it
// as a fixed 8-byte word. Used by INETA_CACHE_STATISTICS.CurrentFileCacheSize and
// .MaximumFileCacheSize ([MS-IRP] 2.2.1.6).
type DWORDLONG = ndr.DWORD64

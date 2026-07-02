package msbrwsa

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// BROWSER_IDENTIFY_HANDLE is a [handle] LPWSTR ([MS-BRWSA] 2.2.2.1): a generic-handle
// alias for a wide string used as the ServerName parameter of I_BrowserrQueryOtherDomains.
type BROWSER_IDENTIFY_HANDLE = ndr.WSTR

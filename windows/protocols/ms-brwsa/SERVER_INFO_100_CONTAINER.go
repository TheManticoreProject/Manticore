package msbrwsa

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_100_CONTAINER contains a count of the entries returned by
// I_BrowserrQueryOtherDomains together with a pointer to the array of entries
// ([MS-BRWSA] 2.2.3.1). Buffer is a [size_is(EntriesRead)] LPSERVER_INFO_100 — a [unique]
// pointer to a conformant array of SERVER_INFO_100 sized by EntriesRead.
type SERVER_INFO_100_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []SERVER_INFO_100 `ndr:"unique,size_is=EntriesRead"`
}

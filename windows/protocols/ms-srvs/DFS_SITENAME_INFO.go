package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// DFS_SITENAME_INFO contains the site name of a server and associated flags
// ([MS-SRVS] 2.2.4.90). SiteName is a [string,unique] WCHAR*.
type DFS_SITENAME_INFO struct {
	SiteFlags ndr.DWORD
	SiteName  ndr.WSTR `ndr:"unique"`
}

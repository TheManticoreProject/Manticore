package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// DFS_SITELIST_INFO contains the list of sites associated with a server
// ([MS-SRVS] 2.2.4.91). Site is an embedded conformant array (not a pointer)
// sized by CSites, so it must be the final field of the structure.
type DFS_SITELIST_INFO struct {
	CSites ndr.DWORD
	Site   []DFS_SITENAME_INFO `ndr:"conformant,size_is=CSites"`
}

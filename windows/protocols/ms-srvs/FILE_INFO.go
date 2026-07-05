package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// FILE_INFO is the [switch_type(unsigned long)] union returned by NetrFileGetInfo
// ([MS-SRVS] 2.2.3.2). Tag carries the discriminant (the level) inline, followed by
// the selected arm. Each arm is a [unique] pointer to the file information struct.
type FILE_INFO struct {
	Tag       ndr.DWORD    `ndr:"switch"`
	FileInfo2 *FILE_INFO_2 `ndr:"case=2,unique"`
	FileInfo3 *FILE_INFO_3 `ndr:"case=3,unique"`
}

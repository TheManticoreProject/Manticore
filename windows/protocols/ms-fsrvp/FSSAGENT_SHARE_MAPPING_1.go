package msfsrvp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// FSSAGENT_SHARE_MAPPING_1 models FSSAGENT_SHARE_MAPPING_1 ([MS-FSRVP] 2.2.3.2): the
// level-1 shadow-copy share mapping. The two IDs are [MS-DTYP] GUIDs (16 octets on the
// wire) — modeled as dtyp.GUID, not windows/guid.GUID whose uint64 tail would marshal to
// 24 octets. ShareNameUNC/ShadowCopyShareName are [unique][string] LPWSTR referents;
// CreationTimestamp is a LONGLONG (FILETIME as a signed 64-bit count).
type FSSAGENT_SHARE_MAPPING_1 struct {
	ShadowCopySetId     dtyp.GUID
	ShadowCopyId        dtyp.GUID
	ShareNameUNC        *ndr.WSTR `ndr:"unique"`
	ShadowCopyShareName *ndr.WSTR `ndr:"unique"`
	CreationTimestamp   int64
}

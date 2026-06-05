package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_POLICY_DOMAIN_EFS_INFO communicates a counted binary EFS policy blob ([MS-LSAD]
// 2.2.4.19). EfsBlob is a [unique] pointer to a conformant byte array sized by
// InfoLength.
type LSAPR_POLICY_DOMAIN_EFS_INFO struct {
	InfoLength ndr.DWORD
	EfsBlob    []byte `ndr:"unique,size_is=InfoLength"`
}

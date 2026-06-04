package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_OBJECT_ATTRIBUTES models LSAPR_OBJECT_ATTRIBUTES ([MS-LSAD] 2.2.2.3). All fields are
// ignored by the server except RootDirectory, which must be NULL, so each is modeled
// as a 4-octet zero field (the four pointer members as NULL referents).
type LSAPR_OBJECT_ATTRIBUTES struct {
	Length                   ndr.DWORD
	RootDirectory            ndr.DWORD // [unique] NULL
	ObjectName               ndr.DWORD // [unique] NULL
	Attributes               ndr.DWORD
	SecurityDescriptor       ndr.DWORD // [unique] NULL
	SecurityQualityOfService ndr.DWORD // [unique] NULL
}

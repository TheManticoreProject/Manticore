package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_OBJECT_ATTRIBUTES models LSAPR_OBJECT_ATTRIBUTES ([MS-LSAD] 2.2.2.3). The four
// pointer members are [unique] pointers that the server ignores (RootDirectory must be
// NULL); they are modeled as nil typed pointers, not as scalar fields, so the codec
// emits a real referent id whose width tracks the transfer syntax — 4 octets under
// NDR20, 8 octets 8-aligned under NDR64 ([MS-RPCE] section 2.2.5). A nil [unique]
// pointer marshals to a zero referent identically in NDR20 to the previous DWORD zero,
// so NDR20 output is unchanged; under NDR64 the previous DWORD model produced a 4-octet
// field where the server expects an 8-octet referent, faulting nca_s_fault_ndr.
type LSAPR_OBJECT_ATTRIBUTES struct {
	Length                   ndr.DWORD
	RootDirectory            *uint32 `ndr:"unique"` // always NULL
	ObjectName               *uint32 `ndr:"unique"` // always NULL
	Attributes               ndr.DWORD
	SecurityDescriptor       *uint32 `ndr:"unique"` // always NULL
	SecurityQualityOfService *uint32 `ndr:"unique"` // always NULL
}

package mslsat

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSAPR_SID_INFORMATION wraps a single SID ([MS-LSAT] 2.2.17). Sid is a [unique]
// pointer to an RPC_SID (the interfaces that use this type declare
// pointer_default(unique)). This type is defined by MS-LSAT and reused by other
// LSA-family protocols (for example, MS-CAPR's LSAPR_WRAPPED_CAPID_SET).
//
// This is the split-layout home for the type. A legacy, byte-identical copy still
// lives inline under the lsarpc interface tree
// (network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures);
// when the rest of the MS-LSAT wire structures are migrated here, that copy should
// be deleted and its users pointed at this package.
type LSAPR_SID_INFORMATION struct {
	Sid *dtyp.RPC_SID `ndr:"unique"`
}

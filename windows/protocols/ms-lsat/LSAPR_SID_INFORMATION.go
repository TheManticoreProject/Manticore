package mslsat

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSAPR_SID_INFORMATION wraps a single SID ([MS-LSAT] 2.2.17). Sid is a [unique]
// pointer to an RPC_SID (the interfaces that use this type declare
// pointer_default(unique)). This type is defined by MS-LSAT and reused by other
// LSA-family protocols (for example, MS-CAPR's LSAPR_WRAPPED_CAPID_SET).
//
// This package is the single canonical home for the MS-LSAT wire structures
// (LSAPR_SID_ENUM_BUFFER, the LSAPR_TRANSLATED_* types, LSAPR_REFERENCED_DOMAIN_LIST, and
// this one); lsarpc's LsarLookupSids*/LsarLookupNames* methods bind to these definitions.
type LSAPR_SID_INFORMATION struct {
	Sid *dtyp.RPC_SID `ndr:"unique"`
}

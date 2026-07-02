package mslsat

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSAPR_SID_INFORMATION wraps a single SID ([MS-LSAT] 2.2.17). Sid is a [unique]
// pointer to an RPC_SID (the interfaces that use this type declare
// pointer_default(unique)). This type is defined by MS-LSAT and reused by other
// LSA-family protocols (for example, MS-CAPR's LSAPR_WRAPPED_CAPID_SET).
//
// This is the intended split-layout home for the type. A byte-identical copy currently
// also lives in the MS-LSAD structures package (windows/protocols/ms-lsad), which is where
// the shared lsarpc structures were relocated when MS-LSAD was migrated to the split
// layout; lsarpc's LsarLookupSids* methods bind to that copy. When the remaining MS-LSAT
// wire structures (LSAPR_SID_ENUM_BUFFER, the LSAPR_TRANSLATED_* types, and this one) are
// consolidated into this package, that duplicate should be deleted and its users pointed
// here so the type has a single canonical home.
type LSAPR_SID_INFORMATION struct {
	Sid *dtyp.RPC_SID `ndr:"unique"`
}

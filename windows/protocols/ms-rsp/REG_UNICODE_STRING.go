package msrsp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// REG_UNICODE_STRING is the counted Unicode string used by the InitShutdown and
// WindowsShutdown interfaces ([MS-RSP] section 2.2.2). It is byte-for-byte the
// [MS-DTYP] RPC_UNICODE_STRING (Length/MaximumLength are BYTE counts; Buffer is a
// [unique] pointer to a conformant-varying wchar array whose maximum_count is
// MaximumLength/2 and actual_count is Length/2), so it is modeled as an alias of the
// shared dtyp type rather than redefined — matching how MS-RRP models its identical
// RRP_UNICODE_STRING. Build with dtyp.NewUnicodeString; read with .String().
type REG_UNICODE_STRING = dtyp.RPC_UNICODE_STRING

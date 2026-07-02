package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_POLICY_AUDIT_EVENTS_INFO contains auditing options ([MS-LSAD] 2.2.4.4).
// EventAuditingOptions is a [unique] pointer to a conformant array of unsigned long
// sized by MaximumAuditEventCount.
type LSAPR_POLICY_AUDIT_EVENTS_INFO struct {
	AuditingMode           uint8
	EventAuditingOptions   []ndr.DWORD `ndr:"unique,size_is=MaximumAuditEventCount"`
	MaximumAuditEventCount ndr.DWORD
}

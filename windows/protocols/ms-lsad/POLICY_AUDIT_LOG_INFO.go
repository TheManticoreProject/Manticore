package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// POLICY_AUDIT_LOG_INFO contains information about the state of the audit log
// ([MS-LSAD] 2.2.4.3).
type POLICY_AUDIT_LOG_INFO struct {
	AuditLogPercentFull            ndr.DWORD
	MaximumLogSize                 ndr.DWORD
	AuditRetentionPeriod           dtyp.LARGE_INTEGER
	AuditLogFullShutdownInProgress uint8
	TimeToShutdown                 dtyp.LARGE_INTEGER
	NextAuditRecordId              ndr.DWORD
}

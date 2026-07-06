package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// POLICY_AUDIT_LOG_INFO contains information about the state of the audit log
// ([MS-LSAD] 2.2.4.3).
type POLICY_AUDIT_LOG_INFO struct {
	AuditLogPercentFull            ndr.DWORD
	MaximumLogSize                 ndr.DWORD
	AuditRetentionPeriod           msdtyp.LARGE_INTEGER
	AuditLogFullShutdownInProgress uint8
	TimeToShutdown                 msdtyp.LARGE_INTEGER
	NextAuditRecordId              ndr.DWORD
}

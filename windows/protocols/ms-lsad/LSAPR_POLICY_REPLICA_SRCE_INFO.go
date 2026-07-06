package mslsad

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// LSAPR_POLICY_REPLICA_SRCE_INFO contains replica source information ([MS-LSAD]
// 2.2.4.9).
type LSAPR_POLICY_REPLICA_SRCE_INFO struct {
	ReplicaSource      msdtyp.RPC_UNICODE_STRING
	ReplicaAccountName msdtyp.RPC_UNICODE_STRING
}

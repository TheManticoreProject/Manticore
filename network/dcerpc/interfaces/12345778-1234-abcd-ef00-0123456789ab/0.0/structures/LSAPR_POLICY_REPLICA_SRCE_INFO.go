package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSAPR_POLICY_REPLICA_SRCE_INFO contains replica source information ([MS-LSAD]
// 2.2.4.9).
type LSAPR_POLICY_REPLICA_SRCE_INFO struct {
	ReplicaSource      dtyp.RPC_UNICODE_STRING
	ReplicaAccountName dtyp.RPC_UNICODE_STRING
}

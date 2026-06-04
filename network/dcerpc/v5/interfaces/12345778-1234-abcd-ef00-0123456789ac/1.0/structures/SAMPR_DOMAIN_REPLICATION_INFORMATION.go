package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_DOMAIN_REPLICATION_INFORMATION contains the replica source node name for a
// domain ([MS-SAMR] 2.2.4.13).
type SAMPR_DOMAIN_REPLICATION_INFORMATION struct {
	ReplicaSourceNodeName dtyp.RPC_UNICODE_STRING
}

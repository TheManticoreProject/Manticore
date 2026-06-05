package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSA_FOREST_TRUST_DOMAIN_INFO identifies a domain within a forest-trust record
// ([MS-LSAD] 2.2.7.23). Sid is a [unique] pointer to an RPC_SID; DnsName and NetbiosName
// are LSA_UNICODE_STRING (the same wire form as RPC_UNICODE_STRING).
type LSA_FOREST_TRUST_DOMAIN_INFO struct {
	Sid         *dtyp.RPC_SID `ndr:"unique"`
	DnsName     dtyp.RPC_UNICODE_STRING
	NetbiosName dtyp.RPC_UNICODE_STRING
}

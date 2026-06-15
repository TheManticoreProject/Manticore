package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSAPR_POLICY_DNS_DOMAIN_INFO contains DNS information about the primary domain
// ([MS-LSAD] 2.2.4.16). Sid is a [unique] pointer to an RPC_SID.
type LSAPR_POLICY_DNS_DOMAIN_INFO struct {
	Name          dtyp.RPC_UNICODE_STRING
	DnsDomainName dtyp.RPC_UNICODE_STRING
	DnsForestName dtyp.RPC_UNICODE_STRING
	DomainGuid    dtyp.GUID
	Sid           *dtyp.RPC_SID `ndr:"unique"`
}

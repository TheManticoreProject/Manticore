package mslsad

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// LSAPR_POLICY_DNS_DOMAIN_INFO contains DNS information about the primary domain
// ([MS-LSAD] 2.2.4.16). Sid is a [unique] pointer to an RPC_SID.
type LSAPR_POLICY_DNS_DOMAIN_INFO struct {
	Name          msdtyp.RPC_UNICODE_STRING
	DnsDomainName msdtyp.RPC_UNICODE_STRING
	DnsForestName msdtyp.RPC_UNICODE_STRING
	DomainGuid    msdtyp.GUID
	Sid           *msdtyp.RPC_SID `ndr:"unique"`
}

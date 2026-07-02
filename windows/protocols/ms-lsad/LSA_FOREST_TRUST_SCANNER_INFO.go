package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSA_FOREST_TRUST_SCANNER_INFO carries the domain identity discovered by the forest-trust
// scanner ([MS-LSAD] 2.2.7.28). DomainSid is a [unique] pointer to an RPC_SID (the
// interface declares pointer_default(unique)); DnsName and NetbiosName are counted
// Unicode strings.
type LSA_FOREST_TRUST_SCANNER_INFO struct {
	DomainSid   *dtyp.RPC_SID `ndr:"unique"`
	DnsName     LSA_UNICODE_STRING
	NetbiosName LSA_UNICODE_STRING
}

package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_TRUSTED_DOMAIN_AUTH_BLOB is an encrypted trust authentication blob ([MS-LSAD]
// 2.2.7.14). AuthBlob is a [unique] pointer to a conformant byte array sized by
// AuthSize.
type LSAPR_TRUSTED_DOMAIN_AUTH_BLOB struct {
	AuthSize ndr.DWORD
	AuthBlob []byte `ndr:"unique,size_is=AuthSize"`
}

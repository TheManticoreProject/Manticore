package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSAPR_TRANSLATED_NAME is the translation of a SID into an account name and domain
// index ([MS-LSAT] 2.2.20). Use is an NDR enum (16-bit on the wire); DomainIndex is a
// signed long.
type LSAPR_TRANSLATED_NAME struct {
	Use         SID_NAME_USE `ndr:"enum"`
	Name        dtyp.RPC_UNICODE_STRING
	DomainIndex int32
}

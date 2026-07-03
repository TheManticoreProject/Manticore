package msrpcl

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// NSI_UUID_VECTOR_T ([MS-RPCL] 2.2) is a counted vector of object UUIDs. Uuid is an
// embedded conformant array (the struct's trailing [size_is(count)] member, so its
// maximum_count is hoisted to the front of the struct) whose elements are [unique]
// pointers to GUIDs (NSI_UUID_P_T).
type NSI_UUID_VECTOR_T struct {
	Count ndr.DWORD
	Uuid  []NSI_UUID_P_T `ndr:"conformant,size_is=Count,elem=unique"`
}

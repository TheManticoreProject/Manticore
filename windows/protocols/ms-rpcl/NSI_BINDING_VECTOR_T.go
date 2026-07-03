package msrpcl

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// NSI_BINDING_VECTOR_T ([MS-RPCL] 2.2) is a counted vector of bindings. Binding is an
// embedded conformant array (the struct's trailing [size_is(count)] member, so its
// maximum_count is hoisted to the front of the struct); its elements are inline
// NSI_BINDING_T values, not pointers.
type NSI_BINDING_VECTOR_T struct {
	Count   ndr.DWORD
	Binding []NSI_BINDING_T `ndr:"conformant,size_is=Count"`
}

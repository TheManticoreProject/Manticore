package msrpcl

import (
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// NSI_UUID_P_T is a [unique] pointer to a GUID ([MS-RPCL] 2.2). It is used both as
// a standalone object-UUID argument and as the [unique]-pointer element type of the
// conformant array in NSI_UUID_VECTOR_T; the pointer framing comes from the ndr tag
// at each use site.
type NSI_UUID_P_T = *guid.GUID

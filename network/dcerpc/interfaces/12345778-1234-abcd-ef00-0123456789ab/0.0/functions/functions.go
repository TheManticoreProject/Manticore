// Package functions implements the method stubs of the lsarpc interface
// (12345778-1234-abcd-ef00-0123456789ab v0.0). Each opnum lives in its own
// NN_MethodName.go file; this file holds the request/response shapes shared across
// more than one method.
//
// The package depends on the interface descriptor (opnums, status codes) and on the
// structures package (the NDR types); neither depends on this one.
package functions

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// handleResponse is the common reply shape: a 20-byte context handle followed by the
// NTSTATUS return value. It is shared by the methods that return an [out] or [in,out]
// LSAPR_HANDLE (LsarOpenPolicy2, LsarOpenPolicy, LsarClose, LsarDeleteObject, the
// Open*/Create* methods).
type handleResponse struct {
	Handle mslsad.LSAPR_HANDLE
	Status ndr.DWORD `ndr:"retval"`
}

// statusResponse is the reply shape for methods with no [out] parameters beyond the
// NTSTATUS return value (the Set*/Add*/Remove*/Delete* methods).
type statusResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

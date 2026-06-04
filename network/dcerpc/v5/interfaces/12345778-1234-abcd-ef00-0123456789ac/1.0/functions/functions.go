// Package functions implements the method stubs of the samr interface
// (12345778-1234-abcd-ef00-0123456789ac v1.0, [MS-SAMR]). Each opnum lives in its own
// NN_MethodName.go file; this file holds the small shapes and helpers shared across more
// than one method.
package functions

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// statusResponse is the reply shape for methods whose only output is the NTSTATUS return
// value (transmitted last, after any deferred referents).
type statusResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// handleResponse is the reply shape for the [in,out] SAMPR_HANDLE close/delete methods:
// the (now zeroed) handle followed by the NTSTATUS.
type handleResponse struct {
	Handle structures.SAMPR_HANDLE
	Status ndr.DWORD `ndr:"retval"`
}

// openHandleResponse is the reply shape for the Open*/Create* methods that return a fresh
// [out] SAMPR_HANDLE plus the NTSTATUS.
type openHandleResponse struct {
	Handle structures.SAMPR_HANDLE
	Status ndr.DWORD `ndr:"retval"`
}

// optWStr returns a [unique], [string] wide-character pointer for s, or nil for the empty
// string. samr's ServerName arguments are [in,unique,string] PSAMPR_SERVER_NAME pointers
// whose NULL form selects "the local server".
func optWStr(s string) *ndr.WSTR {
	if s == "" {
		return nil
	}
	w := ndr.WSTR(s)
	return &w
}

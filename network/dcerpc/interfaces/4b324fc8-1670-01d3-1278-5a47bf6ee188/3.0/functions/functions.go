// Package functions implements the method stubs of the srvsvc interface
// (4b324fc8-1670-01d3-1278-5a47bf6ee188 v3.0, [MS-SRVS]). Each opnum lives in its own
// NN_MethodName.go file; this file holds the small shapes and helpers shared across
// more than one method.
package functions

// IDL source: [MS-SRVS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/77aacc74-f8f9-4b46-b2d8-bfe04a7d9c44
// A fetched copy is kept at ms-srvs.idl in the interface directory.

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// statusResponse is the reply shape for methods whose only output is the NET_API_STATUS
// return value (the return value is transmitted last, after any deferred referents).
type statusResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// optWStr returns a [unique], [string] wide-character pointer for s, or nil for the empty
// string. srvsvc's ServerName (and the optional filter strings) are [in,string,unique]
// pointers whose NULL form selects "the local server" / "no filter".
func optWStr(s string) *ndr.WSTR {
	if s == "" {
		return nil
	}
	w := ndr.WSTR(s)
	return &w
}

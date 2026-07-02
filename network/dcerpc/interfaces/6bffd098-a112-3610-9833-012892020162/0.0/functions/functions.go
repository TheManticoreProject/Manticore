// Package functions implements the method stubs of the browser interface
// (6bffd098-a112-3610-9833-012892020162 v0.0, [MS-BRWSA]). Each opnum lives in its own
// NN_MethodName.go file; this file holds the small helpers shared across methods.
package functions

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// optWStr returns a [unique], [string] wide-character pointer for s, or nil for the empty
// string. The browser interface's ServerName is an [in,string,unique]
// BROWSER_IDENTIFY_HANDLE whose value is ignored on receipt; nil is a NULL pointer.
func optWStr(s string) *ndr.WSTR {
	if s == "" {
		return nil
	}
	w := ndr.WSTR(s)
	return &w
}

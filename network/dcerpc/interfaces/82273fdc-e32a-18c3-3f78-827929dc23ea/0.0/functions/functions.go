// Package functions holds the eventlog ([MS-EVEN]) method stubs, one file per opnum.
// Each stub marshals its [in] parameters into a request stub, invokes the opnum through
// an ndr.Invoker, and unmarshals the response stub. The shared response shapes below are
// used by more than one method.
package functions

// IDL source: [MS-EVEN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-even/0d0bee9c-dac5-46d9-b19b-2087826c02db
// A fetched copy is kept at ms-even.idl in the interface directory.

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even"
)

// statusResponse is the response stub for methods whose only [out] value is the returned
// status (ElfrClearELFW/A, ElfrBackupELFW/A, ElfrChangeNotify).
type statusResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// handleResponse is the response stub for methods that return an IELF_HANDLE context
// handle plus a status: the Open*/Register* methods ([out] IELF_HANDLE*) and the
// ElfrCloseEL/ElfrDeregisterEventSource methods ([in,out] IELF_HANDLE*).
type handleResponse struct {
	LogHandle mseven.IELF_HANDLE
	Status    ndr.DWORD `ndr:"retval"`
}

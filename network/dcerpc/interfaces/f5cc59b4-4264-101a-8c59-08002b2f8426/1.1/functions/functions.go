// Package functions holds the frsrpc ([MS-FRS1]) method stubs, one file per opnum. Each
// stub marshals its [in] parameters into a request stub, invokes the opnum through an
// ndr.Invoker, and unmarshals the response stub. The shared response shape below is used
// by more than one method.
package functions

// IDL source: [MS-FRS1] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-frs1/dd60a0d9-176a-46f4-9904-000172041b92
// A fetched copy is kept at ms-frs1.idl in the interface directory.

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// statusResponse is the response stub for methods whose only [out] value is the returned
// Win32 status (FrsRpcSendCommPkt, FrsRpcVerifyPromotionParent, FrsNOP).
type statusResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

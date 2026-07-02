// Package functions holds the frsrpc ([MS-FRS1]) method stubs, one file per opnum. Each
// stub marshals its [in] parameters into a request stub, invokes the opnum through an
// ndr.Invoker, and unmarshals the response stub. The shared response shape below is used
// by more than one method.
package functions

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// statusResponse is the response stub for methods whose only [out] value is the returned
// Win32 status (FrsRpcSendCommPkt, FrsRpcVerifyPromotionParent, FrsNOP).
type statusResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

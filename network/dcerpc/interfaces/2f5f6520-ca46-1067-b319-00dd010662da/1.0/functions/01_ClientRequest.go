package functions

import (
	"fmt"

	tapsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/2f5f6520-ca46-1067-b319-00dd010662da/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mstrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-trp"
)

// clientRequestRequest carries the [in]/[in,out] parameters of ClientRequest.
//
// pBuffer is a top-level [ref] pointer to a conformant-varying byte array with
// independent bounds ([MS-TRP] 3.2.4.2): its maximum_count comes from lNeededSize (the
// buffer capacity) and its actual_count from plUsedSize (the number of bytes actually
// carried). Modeling it [ref] (not [unique]) suppresses both a referent id and the
// hoisting of maximum_count ahead of the preceding context handle, matching the wire.
type clientRequestRequest struct {
	PhContext mstrp.PCONTEXT_HANDLE_TYPE
	PBuffer   []uint8 `ndr:"ref,varying,size_is=LNeededSize,length_is=PlUsedSize"`
	// LNeededSize and PlUsedSize are the IDL's signed long parameters; they are modeled as
	// unsigned DWORDs because they are non-negative sizes and the NDR codec resolves the
	// array's independent maximum_count/actual_count from these sibling fields (which it
	// reads as unsigned). On the wire a signed long and a DWORD are the same 4 octets.
	LNeededSize ndr.DWORD
	PlUsedSize  ndr.DWORD
}

func (*clientRequestRequest) Opnum() uint16 { return tapsrv.OpnumClientRequest }

// clientRequestResponse carries the [in,out] parameters of ClientRequest. ClientRequest
// is a void method, so there is no return value on the wire: the request status is
// delivered inside the returned pBuffer payload. On decode the array's maximum_count /
// actual_count are read directly from the wire (size_is/length_is siblings are only
// consulted on the marshal path), so the buffer needs only the [ref] conformant-varying
// attributes here and lNeededSize is intentionally absent.
type clientRequestResponse struct {
	PBuffer    []uint8 `ndr:"ref,varying"`
	PlUsedSize ndr.DWORD
}

// ClientRequest calls ClientRequest (opnum 1) ([MS-TRP] 3.2.4.2). The client sends a
// packed TAPI request in pBuffer; the server processes it and returns the (possibly
// updated) buffer in place. lNeededSize is the total capacity of pBuffer and plUsedSize
// the number of valid bytes; both are [in,out]. The returned buffer and used size are
// reported back to the caller. Transport-level failures surface through err; the TAPI
// result itself is encoded within the returned buffer.
func ClientRequest(rpc ndr.Invoker, phContext mstrp.PCONTEXT_HANDLE_TYPE, pBuffer []uint8, lNeededSize ndr.DWORD, plUsedSize ndr.DWORD) (PBuffer []uint8, PlUsedSize ndr.DWORD, err error) {
	req := &clientRequestRequest{
		PhContext:   phContext,
		PBuffer:     pBuffer,
		LNeededSize: lNeededSize,
		PlUsedSize:  plUsedSize,
	}
	var resp clientRequestResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ClientRequest: %w", err)
		return
	}
	PBuffer = resp.PBuffer
	PlUsedSize = resp.PlUsedSize
	return
}

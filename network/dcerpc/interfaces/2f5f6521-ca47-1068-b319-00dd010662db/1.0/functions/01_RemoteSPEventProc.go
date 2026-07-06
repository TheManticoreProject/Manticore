package functions

// IDL source: [MS-TRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-trp/e86aca98-76e9-4515-9de1-2cadb9084a2b
// A fetched copy is kept at ms-trp.idl in the interface directory.

import (
	"fmt"

	remotesp "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/2f5f6521-ca47-1068-b319-00dd010662db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mstrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-trp"
)

// remoteSPEventProcRequest carries the [in] parameters of RemoteSPEventProc. pBuffer is a
// top-level conformant-varying byte array ([MS-TRP] 3.1.4.2) whose maximum_count and
// actual_count both come from lSize; it is transmitted inline (no referent id).
type remoteSPEventProcRequest struct {
	PhContext mstrp.PCONTEXT_HANDLE_TYPE2
	PBuffer   []uint8 `ndr:"ref,varying,size_is=LSize,length_is=LSize"`
	// LSize is the IDL's signed long; modeled as an unsigned DWORD (a non-negative size,
	// same 4 octets on the wire) so the NDR codec derives the array's maximum_count and
	// actual_count from it and keeps the count consistent with the transmitted elements.
	LSize ndr.DWORD
}

func (*remoteSPEventProcRequest) Opnum() uint16 { return remotesp.OpnumRemoteSPEventProc }

// remoteSPEventProcResponse is empty: RemoteSPEventProc is a void method with no [out]
// parameters, so the response stub carries no data on the wire.
type remoteSPEventProcResponse struct {
}

// RemoteSPEventProc calls RemoteSPEventProc (opnum 1) ([MS-TRP] 3.1.4.2). The telephony
// server invokes it on the client to push an asynchronous TAPI event: pBuffer holds the
// packed event data of lSize bytes for the session identified by phContext. It returns no
// value; transport-level failures surface through err.
func RemoteSPEventProc(rpc ndr.Invoker, phContext mstrp.PCONTEXT_HANDLE_TYPE2, pBuffer []uint8, lSize ndr.DWORD) (err error) {
	req := &remoteSPEventProcRequest{
		PhContext: phContext,
		PBuffer:   pBuffer,
		LSize:     lSize,
	}
	var resp remoteSPEventProcResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RemoteSPEventProc: %w", err)
		return
	}
	return
}

package functions

// IDL source: [MS-FRS2] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-frs2/39bd498b-2a94-41b7-914e-10921d543912
// A fetched copy is kept at ms-frs2.idl in the interface directory.

import (
	"fmt"

	FrsTransport "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/897e2e5f-93f3-4376-9c9c-fd2277495c27/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfrs2 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs2"
)

// rdcGetFileDataAsyncRequest carries the [in] parameters of RdcGetFileDataAsync.
type rdcGetFileDataAsyncRequest struct {
	ServerContext msfrs2.PFRS_SERVER_CONTEXT
}

func (*rdcGetFileDataAsyncRequest) Opnum() uint16 { return FrsTransport.OpnumRdcGetFileDataAsync }

// rdcGetFileDataAsyncResponse carries the [out] parameters and return value of RdcGetFileDataAsync.
type rdcGetFileDataAsyncResponse struct {
	BytePipe msfrs2.BYTE_PIPE `ndr:"pipe"`
	Status   ndr.DWORD        `ndr:"retval"`
}

// RdcGetFileDataAsync calls RdcGetFileDataAsync (opnum 16) ([MS-FRS2] — verify the parameter
// modeling and status handling).
func RdcGetFileDataAsync(rpc ndr.Invoker, serverContext msfrs2.PFRS_SERVER_CONTEXT) (BytePipe msfrs2.BYTE_PIPE, err error) {
	req := &rdcGetFileDataAsyncRequest{
		ServerContext: serverContext,
	}
	var resp rdcGetFileDataAsyncResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RdcGetFileDataAsync: %w", err)
		return
	}
	BytePipe = resp.BytePipe
	if uint32(resp.Status) != FrsTransport.StatusSuccess {
		err = fmt.Errorf("RdcGetFileDataAsync failed: %s", FrsTransport.StatusString(uint32(resp.Status)))
	}
	return
}

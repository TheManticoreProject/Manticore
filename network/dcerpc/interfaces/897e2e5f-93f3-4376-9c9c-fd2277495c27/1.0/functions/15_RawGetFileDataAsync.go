package functions

import (
	"fmt"

	FrsTransport "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/897e2e5f-93f3-4376-9c9c-fd2277495c27/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfrs2 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs2"
)

// rawGetFileDataAsyncRequest carries the [in] parameters of RawGetFileDataAsync.
type rawGetFileDataAsyncRequest struct {
	ServerContext msfrs2.PFRS_SERVER_CONTEXT
}

func (*rawGetFileDataAsyncRequest) Opnum() uint16 { return FrsTransport.OpnumRawGetFileDataAsync }

// rawGetFileDataAsyncResponse carries the [out] parameters and return value of RawGetFileDataAsync.
type rawGetFileDataAsyncResponse struct {
	BytePipe msfrs2.BYTE_PIPE `ndr:"pipe"`
	Status   ndr.DWORD        `ndr:"retval"`
}

// RawGetFileDataAsync calls RawGetFileDataAsync (opnum 15) ([MS-FRS2] — verify the parameter
// modeling and status handling).
func RawGetFileDataAsync(rpc ndr.Invoker, serverContext msfrs2.PFRS_SERVER_CONTEXT) (BytePipe msfrs2.BYTE_PIPE, err error) {
	req := &rawGetFileDataAsyncRequest{
		ServerContext: serverContext,
	}
	var resp rawGetFileDataAsyncResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RawGetFileDataAsync: %w", err)
		return
	}
	BytePipe = resp.BytePipe
	if uint32(resp.Status) != FrsTransport.StatusSuccess {
		err = fmt.Errorf("RawGetFileDataAsync failed: %s", FrsTransport.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	FrsTransport "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/897e2e5f-93f3-4376-9c9c-fd2277495c27/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfrs2 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs2"
)

// rdcGetFileDataRequest carries the [in] parameters of RdcGetFileData.
type rdcGetFileDataRequest struct {
	ServerContext msfrs2.PFRS_SERVER_CONTEXT
	BufferSize    ndr.DWORD
}

func (*rdcGetFileDataRequest) Opnum() uint16 { return FrsTransport.OpnumRdcGetFileData }

// rdcGetFileDataResponse carries the [out] parameters and return value of RdcGetFileData.
type rdcGetFileDataResponse struct {
	DataBuffer   []uint8 `ndr:"ref,size_is=BufferSize,varying"`
	SizeReturned ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// RdcGetFileData calls RdcGetFileData (opnum 11) ([MS-FRS2] — verify the parameter
// modeling and status handling).
func RdcGetFileData(rpc ndr.Invoker, serverContext msfrs2.PFRS_SERVER_CONTEXT, bufferSize ndr.DWORD) (DataBuffer []uint8, SizeReturned ndr.DWORD, err error) {
	req := &rdcGetFileDataRequest{
		ServerContext: serverContext,
		BufferSize:    bufferSize,
	}
	var resp rdcGetFileDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RdcGetFileData: %w", err)
		return
	}
	DataBuffer = resp.DataBuffer
	SizeReturned = resp.SizeReturned
	if uint32(resp.Status) != FrsTransport.StatusSuccess {
		err = fmt.Errorf("RdcGetFileData failed: %s", FrsTransport.StatusString(uint32(resp.Status)))
	}
	return
}

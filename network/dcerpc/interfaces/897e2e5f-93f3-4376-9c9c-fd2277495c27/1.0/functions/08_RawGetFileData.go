package functions

import (
	"fmt"

	FrsTransport "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/897e2e5f-93f3-4376-9c9c-fd2277495c27/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfrs2 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs2"
)

// rawGetFileDataRequest carries the [in] parameters of RawGetFileData.
type rawGetFileDataRequest struct {
	ServerContext msfrs2.PFRS_SERVER_CONTEXT
	BufferSize    ndr.DWORD
}

func (*rawGetFileDataRequest) Opnum() uint16 { return FrsTransport.OpnumRawGetFileData }

// rawGetFileDataResponse carries the [out] parameters and return value of RawGetFileData.
type rawGetFileDataResponse struct {
	ServerContext msfrs2.PFRS_SERVER_CONTEXT
	DataBuffer    []uint8 `ndr:"ref,size_is=BufferSize,varying"`
	SizeRead      ndr.DWORD
	IsEndOfFile   int32
	Status        ndr.DWORD `ndr:"retval"`
}

// RawGetFileData calls RawGetFileData (opnum 8) ([MS-FRS2] — verify the parameter
// modeling and status handling).
func RawGetFileData(rpc ndr.Invoker, serverContext msfrs2.PFRS_SERVER_CONTEXT, bufferSize ndr.DWORD) (ServerContext msfrs2.PFRS_SERVER_CONTEXT, DataBuffer []uint8, SizeRead ndr.DWORD, IsEndOfFile int32, err error) {
	req := &rawGetFileDataRequest{
		ServerContext: serverContext,
		BufferSize:    bufferSize,
	}
	var resp rawGetFileDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RawGetFileData: %w", err)
		return
	}
	ServerContext = resp.ServerContext
	DataBuffer = resp.DataBuffer
	SizeRead = resp.SizeRead
	IsEndOfFile = resp.IsEndOfFile
	if uint32(resp.Status) != FrsTransport.StatusSuccess {
		err = fmt.Errorf("RawGetFileData failed: %s", FrsTransport.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	FrsTransport "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/897e2e5f-93f3-4376-9c9c-fd2277495c27/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfrs2 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs2"
)

// initializeFileTransferAsyncRequest carries the [in] parameters of InitializeFileTransferAsync.
type initializeFileTransferAsyncRequest struct {
	ConnectionId  msfrs2.FRS_CONNECTION_ID
	FrsUpdate     msfrs2.FRS_UPDATE
	RdcDesired    int32
	StagingPolicy msfrs2.FRS_REQUESTED_STAGING_POLICY
	BufferSize    ndr.DWORD
}

func (*initializeFileTransferAsyncRequest) Opnum() uint16 {
	return FrsTransport.OpnumInitializeFileTransferAsync
}

// initializeFileTransferAsyncResponse carries the [out] parameters and return value of InitializeFileTransferAsync.
type initializeFileTransferAsyncResponse struct {
	FrsUpdate     msfrs2.FRS_UPDATE
	StagingPolicy msfrs2.FRS_REQUESTED_STAGING_POLICY
	// An NDR context handle (ndr_context_handle: unsigned long + GUID) is 4-octet aligned
	// ([MS-RPCE] 2.2.1.2.5), but PFRS_SERVER_CONTEXT is modeled as a bare [20]byte
	// (alignment 1) and here it directly follows the 16-bit StagingPolicy enum. Without
	// this override the codec would place the handle 2 bytes early, desyncing the whole
	// response tail. This is the only FrsTransport method where a context handle trails a
	// sub-4-aligned field; elsewhere it is the first response field or follows 4/8-aligned
	// data and lands aligned without help.
	ServerContext msfrs2.PFRS_SERVER_CONTEXT `ndr:"align=4"`
	RdcFileInfo   *msfrs2.FRS_RDC_FILEINFO   `ndr:"unique"`
	DataBuffer    []uint8                    `ndr:"ref,size_is=BufferSize,varying"`
	SizeRead      ndr.DWORD
	IsEndOfFile   int32
	Status        ndr.DWORD `ndr:"retval"`
}

// InitializeFileTransferAsync calls InitializeFileTransferAsync (opnum 13) ([MS-FRS2] — verify the parameter
// modeling and status handling).
func InitializeFileTransferAsync(rpc ndr.Invoker, connectionId msfrs2.FRS_CONNECTION_ID, frsUpdate msfrs2.FRS_UPDATE, rdcDesired int32, stagingPolicy msfrs2.FRS_REQUESTED_STAGING_POLICY, bufferSize ndr.DWORD) (FrsUpdate msfrs2.FRS_UPDATE, StagingPolicy msfrs2.FRS_REQUESTED_STAGING_POLICY, ServerContext msfrs2.PFRS_SERVER_CONTEXT, RdcFileInfo *msfrs2.FRS_RDC_FILEINFO, DataBuffer []uint8, SizeRead ndr.DWORD, IsEndOfFile int32, err error) {
	req := &initializeFileTransferAsyncRequest{
		ConnectionId:  connectionId,
		FrsUpdate:     frsUpdate,
		RdcDesired:    rdcDesired,
		StagingPolicy: stagingPolicy,
		BufferSize:    bufferSize,
	}
	var resp initializeFileTransferAsyncResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("InitializeFileTransferAsync: %w", err)
		return
	}
	FrsUpdate = resp.FrsUpdate
	StagingPolicy = resp.StagingPolicy
	ServerContext = resp.ServerContext
	RdcFileInfo = resp.RdcFileInfo
	DataBuffer = resp.DataBuffer
	SizeRead = resp.SizeRead
	IsEndOfFile = resp.IsEndOfFile
	if uint32(resp.Status) != FrsTransport.StatusSuccess {
		err = fmt.Errorf("InitializeFileTransferAsync failed: %s", FrsTransport.StatusString(uint32(resp.Status)))
	}
	return
}

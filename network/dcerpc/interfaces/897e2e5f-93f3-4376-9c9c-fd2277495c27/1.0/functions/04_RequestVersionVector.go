package functions

import (
	"fmt"

	FrsTransport "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/897e2e5f-93f3-4376-9c9c-fd2277495c27/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfrs2 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs2"
)

// requestVersionVectorRequest carries the [in] parameters of RequestVersionVector.
type requestVersionVectorRequest struct {
	SequenceNumber ndr.DWORD
	ConnectionId   msfrs2.FRS_CONNECTION_ID
	ContentSetId   msfrs2.FRS_CONTENT_SET_ID
	RequestType    msfrs2.VERSION_REQUEST_TYPE
	ChangeType     msfrs2.VERSION_CHANGE_TYPE
	VvGeneration   uint64
}

func (*requestVersionVectorRequest) Opnum() uint16 { return FrsTransport.OpnumRequestVersionVector }

// requestVersionVectorResponse carries the [out] parameters and return value of RequestVersionVector.
type requestVersionVectorResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RequestVersionVector calls RequestVersionVector (opnum 4) ([MS-FRS2] — verify the parameter
// modeling and status handling).
func RequestVersionVector(rpc ndr.Invoker, sequenceNumber ndr.DWORD, connectionId msfrs2.FRS_CONNECTION_ID, contentSetId msfrs2.FRS_CONTENT_SET_ID, requestType msfrs2.VERSION_REQUEST_TYPE, changeType msfrs2.VERSION_CHANGE_TYPE, vvGeneration uint64) (err error) {
	req := &requestVersionVectorRequest{
		SequenceNumber: sequenceNumber,
		ConnectionId:   connectionId,
		ContentSetId:   contentSetId,
		RequestType:    requestType,
		ChangeType:     changeType,
		VvGeneration:   vvGeneration,
	}
	var resp requestVersionVectorResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RequestVersionVector: %w", err)
		return
	}
	if uint32(resp.Status) != FrsTransport.StatusSuccess {
		err = fmt.Errorf("RequestVersionVector failed: %s", FrsTransport.StatusString(uint32(resp.Status)))
	}
	return
}

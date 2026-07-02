package functions

import (
	"fmt"

	FrsTransport "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/897e2e5f-93f3-4376-9c9c-fd2277495c27/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfrs2 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs2"
)

// requestRecordsRequest carries the [in] parameters of RequestRecords.
type requestRecordsRequest struct {
	ConnectionId msfrs2.FRS_CONNECTION_ID
	ContentSetId msfrs2.FRS_CONTENT_SET_ID
	UidDbGuid    msfrs2.FRS_DATABASE_ID
	UidVersion   uint64
	MaxRecords   ndr.DWORD
}

func (*requestRecordsRequest) Opnum() uint16 { return FrsTransport.OpnumRequestRecords }

// requestRecordsResponse carries the [out] parameters and return value of RequestRecords.
type requestRecordsResponse struct {
	MaxRecords        ndr.DWORD
	NumRecords        ndr.DWORD
	NumBytes          ndr.DWORD
	CompressedRecords []byte `ndr:"unique,size_is=NumBytes"`
	RecordsStatus     msfrs2.RECORDS_STATUS
	Status            ndr.DWORD `ndr:"retval"`
}

// RequestRecords calls RequestRecords (opnum 6) ([MS-FRS2] — verify the parameter
// modeling and status handling).
func RequestRecords(rpc ndr.Invoker, connectionId msfrs2.FRS_CONNECTION_ID, contentSetId msfrs2.FRS_CONTENT_SET_ID, uidDbGuid msfrs2.FRS_DATABASE_ID, uidVersion uint64, maxRecords ndr.DWORD) (MaxRecords ndr.DWORD, NumRecords ndr.DWORD, NumBytes ndr.DWORD, CompressedRecords []byte, RecordsStatus msfrs2.RECORDS_STATUS, err error) {
	req := &requestRecordsRequest{
		ConnectionId: connectionId,
		ContentSetId: contentSetId,
		UidDbGuid:    uidDbGuid,
		UidVersion:   uidVersion,
		MaxRecords:   maxRecords,
	}
	var resp requestRecordsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RequestRecords: %w", err)
		return
	}
	MaxRecords = resp.MaxRecords
	NumRecords = resp.NumRecords
	NumBytes = resp.NumBytes
	CompressedRecords = resp.CompressedRecords
	RecordsStatus = resp.RecordsStatus
	if uint32(resp.Status) != FrsTransport.StatusSuccess {
		err = fmt.Errorf("RequestRecords failed: %s", FrsTransport.StatusString(uint32(resp.Status)))
	}
	return
}

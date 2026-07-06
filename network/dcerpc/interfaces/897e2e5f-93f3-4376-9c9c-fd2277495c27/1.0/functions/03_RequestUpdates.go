package functions

import (
	"fmt"

	FrsTransport "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/897e2e5f-93f3-4376-9c9c-fd2277495c27/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	msfrs2 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs2"
)

// requestUpdatesRequest carries the [in] parameters of RequestUpdates.
type requestUpdatesRequest struct {
	ConnectionId           msfrs2.FRS_CONNECTION_ID
	ContentSetId           msfrs2.FRS_CONTENT_SET_ID
	CreditsAvailable       ndr.DWORD
	HashRequested          int32
	UpdateRequestType      msfrs2.UPDATE_REQUEST_TYPE
	VersionVectorDiffCount ndr.DWORD
	VersionVectorDiff      []msfrs2.FRS_VERSION_VECTOR `ndr:"ref,size_is=VersionVectorDiffCount"`
}

func (*requestUpdatesRequest) Opnum() uint16 { return FrsTransport.OpnumRequestUpdates }

// requestUpdatesResponse carries the [out] parameters and return value of RequestUpdates.
type requestUpdatesResponse struct {
	FrsUpdate    []msfrs2.FRS_UPDATE `ndr:"ref,size_is=CreditsAvailable,varying"`
	UpdateCount  ndr.DWORD
	UpdateStatus msfrs2.UPDATE_STATUS
	GvsnDbGuid   msdtyp.GUID
	GvsnVersion  uint64
	Status       ndr.DWORD `ndr:"retval"`
}

// RequestUpdates calls RequestUpdates (opnum 3) ([MS-FRS2] — verify the parameter
// modeling and status handling).
func RequestUpdates(rpc ndr.Invoker, connectionId msfrs2.FRS_CONNECTION_ID, contentSetId msfrs2.FRS_CONTENT_SET_ID, creditsAvailable ndr.DWORD, hashRequested int32, updateRequestType msfrs2.UPDATE_REQUEST_TYPE, versionVectorDiffCount ndr.DWORD, versionVectorDiff []msfrs2.FRS_VERSION_VECTOR) (FrsUpdate []msfrs2.FRS_UPDATE, UpdateCount ndr.DWORD, UpdateStatus msfrs2.UPDATE_STATUS, GvsnDbGuid msdtyp.GUID, GvsnVersion uint64, err error) {
	req := &requestUpdatesRequest{
		ConnectionId:           connectionId,
		ContentSetId:           contentSetId,
		CreditsAvailable:       creditsAvailable,
		HashRequested:          hashRequested,
		UpdateRequestType:      updateRequestType,
		VersionVectorDiffCount: versionVectorDiffCount,
		VersionVectorDiff:      versionVectorDiff,
	}
	var resp requestUpdatesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RequestUpdates: %w", err)
		return
	}
	FrsUpdate = resp.FrsUpdate
	UpdateCount = resp.UpdateCount
	UpdateStatus = resp.UpdateStatus
	GvsnDbGuid = resp.GvsnDbGuid
	GvsnVersion = resp.GvsnVersion
	if uint32(resp.Status) != FrsTransport.StatusSuccess {
		err = fmt.Errorf("RequestUpdates failed: %s", FrsTransport.StatusString(uint32(resp.Status)))
	}
	return
}

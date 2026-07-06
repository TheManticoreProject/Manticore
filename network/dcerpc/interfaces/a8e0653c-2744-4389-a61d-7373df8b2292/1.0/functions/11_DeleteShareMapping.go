package functions

import (
	"fmt"

	FileServerVssAgent "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/a8e0653c-2744-4389-a61d-7373df8b2292/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// deleteShareMappingRequest carries the [in] parameters of DeleteShareMapping.
type deleteShareMappingRequest struct {
	ShadowCopySetId msdtyp.GUID
	ShadowCopyId    msdtyp.GUID
	ShareName       ndr.WSTR
}

func (*deleteShareMappingRequest) Opnum() uint16 { return FileServerVssAgent.OpnumDeleteShareMapping }

// deleteShareMappingResponse carries the [out] parameters and return value of DeleteShareMapping.
type deleteShareMappingResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// DeleteShareMapping calls DeleteShareMapping (opnum 11) ([MS-FSRVP] — verify the parameter
// modeling and status handling).
func DeleteShareMapping(rpc ndr.Invoker, shadowCopySetId guid.GUID, shadowCopyId guid.GUID, shareName ndr.WSTR) (err error) {
	req := &deleteShareMappingRequest{
		ShadowCopySetId: msdtyp.NewGUID(shadowCopySetId),
		ShadowCopyId:    msdtyp.NewGUID(shadowCopyId),
		ShareName:       shareName,
	}
	var resp deleteShareMappingResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("DeleteShareMapping: %w", err)
		return
	}
	if uint32(resp.Status) != FileServerVssAgent.StatusSuccess {
		err = fmt.Errorf("DeleteShareMapping failed: %s", FileServerVssAgent.StatusString(uint32(resp.Status)))
	}
	return
}

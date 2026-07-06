package functions

import (
	"fmt"

	FileServerVssAgent "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/a8e0653c-2744-4389-a61d-7373df8b2292/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	msfsrvp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fsrvp"
)

// getShareMappingRequest carries the [in] parameters of GetShareMapping.
type getShareMappingRequest struct {
	ShadowCopyId    msdtyp.GUID
	ShadowCopySetId msdtyp.GUID
	ShareName       ndr.WSTR
	Level           ndr.DWORD
}

func (*getShareMappingRequest) Opnum() uint16 { return FileServerVssAgent.OpnumGetShareMapping }

// getShareMappingResponse carries the [out] parameters and return value of GetShareMapping.
type getShareMappingResponse struct {
	ShareMapping msfsrvp.FSSAGENT_SHARE_MAPPING
	Status       ndr.DWORD `ndr:"retval"`
}

// GetShareMapping calls GetShareMapping (opnum 10) ([MS-FSRVP] — verify the parameter
// modeling and status handling).
func GetShareMapping(rpc ndr.Invoker, shadowCopyId guid.GUID, shadowCopySetId guid.GUID, shareName ndr.WSTR, level ndr.DWORD) (ShareMapping msfsrvp.FSSAGENT_SHARE_MAPPING, err error) {
	req := &getShareMappingRequest{
		ShadowCopyId:    msdtyp.NewGUID(shadowCopyId),
		ShadowCopySetId: msdtyp.NewGUID(shadowCopySetId),
		ShareName:       shareName,
		Level:           level,
	}
	var resp getShareMappingResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("GetShareMapping: %w", err)
		return
	}
	ShareMapping = resp.ShareMapping
	if uint32(resp.Status) != FileServerVssAgent.StatusSuccess {
		err = fmt.Errorf("GetShareMapping failed: %s", FileServerVssAgent.StatusString(uint32(resp.Status)))
	}
	return
}

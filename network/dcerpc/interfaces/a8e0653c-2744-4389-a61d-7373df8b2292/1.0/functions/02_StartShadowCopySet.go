package functions

import (
	"fmt"

	FileServerVssAgent "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/a8e0653c-2744-4389-a61d-7373df8b2292/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// startShadowCopySetRequest carries the [in] parameters of StartShadowCopySet.
type startShadowCopySetRequest struct {
	ClientShadowCopySetId msdtyp.GUID
}

func (*startShadowCopySetRequest) Opnum() uint16 { return FileServerVssAgent.OpnumStartShadowCopySet }

// startShadowCopySetResponse carries the [out] parameters and return value of StartShadowCopySet.
type startShadowCopySetResponse struct {
	PShadowCopySetId msdtyp.GUID
	Status           ndr.DWORD `ndr:"retval"`
}

// StartShadowCopySet calls StartShadowCopySet (opnum 2) ([MS-FSRVP] — verify the parameter
// modeling and status handling).
func StartShadowCopySet(rpc ndr.Invoker, clientShadowCopySetId guid.GUID) (PShadowCopySetId guid.GUID, err error) {
	req := &startShadowCopySetRequest{
		ClientShadowCopySetId: msdtyp.NewGUID(clientShadowCopySetId),
	}
	var resp startShadowCopySetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("StartShadowCopySet: %w", err)
		return
	}
	PShadowCopySetId = resp.PShadowCopySetId.GUID()
	if uint32(resp.Status) != FileServerVssAgent.StatusSuccess {
		err = fmt.Errorf("StartShadowCopySet failed: %s", FileServerVssAgent.StatusString(uint32(resp.Status)))
	}
	return
}

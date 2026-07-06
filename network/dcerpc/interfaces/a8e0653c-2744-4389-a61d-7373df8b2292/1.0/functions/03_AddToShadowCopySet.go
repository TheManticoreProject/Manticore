package functions

import (
	"fmt"

	FileServerVssAgent "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/a8e0653c-2744-4389-a61d-7373df8b2292/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// addToShadowCopySetRequest carries the [in] parameters of AddToShadowCopySet.
type addToShadowCopySetRequest struct {
	ClientShadowCopyId msdtyp.GUID
	ShadowCopySetId    msdtyp.GUID
	ShareName          ndr.WSTR
}

func (*addToShadowCopySetRequest) Opnum() uint16 { return FileServerVssAgent.OpnumAddToShadowCopySet }

// addToShadowCopySetResponse carries the [out] parameters and return value of AddToShadowCopySet.
type addToShadowCopySetResponse struct {
	PShadowCopyId msdtyp.GUID
	Status        ndr.DWORD `ndr:"retval"`
}

// AddToShadowCopySet calls AddToShadowCopySet (opnum 3) ([MS-FSRVP] — verify the parameter
// modeling and status handling).
func AddToShadowCopySet(rpc ndr.Invoker, clientShadowCopyId guid.GUID, shadowCopySetId guid.GUID, shareName ndr.WSTR) (PShadowCopyId guid.GUID, err error) {
	req := &addToShadowCopySetRequest{
		ClientShadowCopyId: msdtyp.NewGUID(clientShadowCopyId),
		ShadowCopySetId:    msdtyp.NewGUID(shadowCopySetId),
		ShareName:          shareName,
	}
	var resp addToShadowCopySetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("AddToShadowCopySet: %w", err)
		return
	}
	PShadowCopyId = resp.PShadowCopyId.GUID()
	if uint32(resp.Status) != FileServerVssAgent.StatusSuccess {
		err = fmt.Errorf("AddToShadowCopySet failed: %s", FileServerVssAgent.StatusString(uint32(resp.Status)))
	}
	return
}

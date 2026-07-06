package functions

// IDL source: [MS-FSRVP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fsrvp/23382633-78f1-419e-bad0-699dff0c6ef1
// A fetched copy is kept at ms-fsrvp.idl in the interface directory.

import (
	"fmt"

	FileServerVssAgent "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/a8e0653c-2744-4389-a61d-7373df8b2292/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// exposeShadowCopySetRequest carries the [in] parameters of ExposeShadowCopySet.
type exposeShadowCopySetRequest struct {
	ShadowCopySetId       msdtyp.GUID
	TimeOutInMilliseconds ndr.DWORD
}

func (*exposeShadowCopySetRequest) Opnum() uint16 { return FileServerVssAgent.OpnumExposeShadowCopySet }

// exposeShadowCopySetResponse carries the [out] parameters and return value of ExposeShadowCopySet.
type exposeShadowCopySetResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// ExposeShadowCopySet calls ExposeShadowCopySet (opnum 5) ([MS-FSRVP] — verify the parameter
// modeling and status handling).
func ExposeShadowCopySet(rpc ndr.Invoker, shadowCopySetId guid.GUID, timeOutInMilliseconds ndr.DWORD) (err error) {
	req := &exposeShadowCopySetRequest{
		ShadowCopySetId:       msdtyp.NewGUID(shadowCopySetId),
		TimeOutInMilliseconds: timeOutInMilliseconds,
	}
	var resp exposeShadowCopySetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ExposeShadowCopySet: %w", err)
		return
	}
	if uint32(resp.Status) != FileServerVssAgent.StatusSuccess {
		err = fmt.Errorf("ExposeShadowCopySet failed: %s", FileServerVssAgent.StatusString(uint32(resp.Status)))
	}
	return
}

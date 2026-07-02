package functions

import (
	"fmt"

	lsacap "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/afc07e2e-311c-4435-808c-c483ffeec7c9/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscapr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-capr"
)

// lsarGetAvailableCAPIDsRequest carries the [in] parameters of LsarGetAvailableCAPIDs.
type lsarGetAvailableCAPIDsRequest struct {
}

func (*lsarGetAvailableCAPIDsRequest) Opnum() uint16 { return lsacap.OpnumLsarGetAvailableCAPIDs }

// lsarGetAvailableCAPIDsResponse carries the [out] parameters and return value of LsarGetAvailableCAPIDs.
type lsarGetAvailableCAPIDsResponse struct {
	WrappedCAPIDs mscapr.LSAPR_WRAPPED_CAPID_SET
	Status        ndr.DWORD `ndr:"retval"`
}

// LsarGetAvailableCAPIDs calls LsarGetAvailableCAPIDs (opnum 0) ([MS-CAPR] 3.1.4.1).
// It returns the set of CAPIDs (the SIDs of the central access policy objects)
// deployed on the remote machine. The single [out] LSAPR_WRAPPED_CAPID_SET* is a
// [ref] pointer, so it is modeled as an inline value in the response.
func LsarGetAvailableCAPIDs(rpc ndr.Invoker) (WrappedCAPIDs mscapr.LSAPR_WRAPPED_CAPID_SET, err error) {
	req := &lsarGetAvailableCAPIDsRequest{}
	var resp lsarGetAvailableCAPIDsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("LsarGetAvailableCAPIDs: %w", err)
		return
	}
	WrappedCAPIDs = resp.WrappedCAPIDs
	if uint32(resp.Status) != lsacap.StatusSuccess {
		err = fmt.Errorf("LsarGetAvailableCAPIDs failed: %s", lsacap.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_GetPersonalCoverPagesOptionRequest carries the [in] parameters of FAX_GetPersonalCoverPagesOption.
type fAX_GetPersonalCoverPagesOptionRequest struct {
}

func (*fAX_GetPersonalCoverPagesOptionRequest) Opnum() uint16 {
	return fax.OpnumFAX_GetPersonalCoverPagesOption
}

// fAX_GetPersonalCoverPagesOptionResponse carries the [out] parameters and return value of FAX_GetPersonalCoverPagesOption.
type fAX_GetPersonalCoverPagesOptionResponse struct {
	LpbPersonalCPAllowed ndr.BOOL
	Status               ndr.DWORD `ndr:"retval"`
}

// FAX_GetPersonalCoverPagesOption calls FAX_GetPersonalCoverPagesOption (opnum 40) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetPersonalCoverPagesOption(rpc ndr.Invoker) (LpbPersonalCPAllowed ndr.BOOL, err error) {
	req := &fAX_GetPersonalCoverPagesOptionRequest{}
	var resp fAX_GetPersonalCoverPagesOptionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetPersonalCoverPagesOption: %w", err)
		return
	}
	LpbPersonalCPAllowed = resp.LpbPersonalCPAllowed
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetPersonalCoverPagesOption failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

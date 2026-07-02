package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_AccessCheckRequest carries the [in] parameters of FAX_AccessCheck.
type fAX_AccessCheckRequest struct {
	AccessMask ndr.DWORD
	LpdwRights *ndr.DWORD `ndr:"unique"`
}

func (*fAX_AccessCheckRequest) Opnum() uint16 { return fax.OpnumFAX_AccessCheck }

// fAX_AccessCheckResponse carries the [out] parameters and return value of FAX_AccessCheck.
type fAX_AccessCheckResponse struct {
	PfAccess   ndr.BOOL
	LpdwRights *ndr.DWORD `ndr:"unique"`
	Status     ndr.DWORD  `ndr:"retval"`
}

// FAX_AccessCheck calls FAX_AccessCheck (opnum 25) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_AccessCheck(rpc ndr.Invoker, accessMask ndr.DWORD, lpdwRights *ndr.DWORD) (PfAccess ndr.BOOL, LpdwRights *ndr.DWORD, err error) {
	req := &fAX_AccessCheckRequest{
		AccessMask: accessMask,
		LpdwRights: lpdwRights,
	}
	var resp fAX_AccessCheckResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_AccessCheck: %w", err)
		return
	}
	PfAccess = resp.PfAccess
	LpdwRights = resp.LpdwRights
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_AccessCheck failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

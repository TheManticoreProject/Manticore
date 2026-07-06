package functions

// IDL source: [MS-FAX] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fax/0907310f-0693-47e7-a6cb-3e599c89a1dd
// A fetched copy is kept at ms-fax.idl in the interface directory.

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_GetServerActivityRequest carries the [in] parameters of FAX_GetServerActivity.
type fAX_GetServerActivityRequest struct {
	PServerActivity msfax.FAX_SERVER_ACTIVITY
}

func (*fAX_GetServerActivityRequest) Opnum() uint16 { return fax.OpnumFAX_GetServerActivity }

// fAX_GetServerActivityResponse carries the [out] parameters and return value of FAX_GetServerActivity.
type fAX_GetServerActivityResponse struct {
	PServerActivity msfax.FAX_SERVER_ACTIVITY
	Status          ndr.DWORD `ndr:"retval"`
}

// FAX_GetServerActivity calls FAX_GetServerActivity (opnum 76) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetServerActivity(rpc ndr.Invoker, pServerActivity msfax.FAX_SERVER_ACTIVITY) (PServerActivity msfax.FAX_SERVER_ACTIVITY, err error) {
	req := &fAX_GetServerActivityRequest{
		PServerActivity: pServerActivity,
	}
	var resp fAX_GetServerActivityResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetServerActivity: %w", err)
		return
	}
	PServerActivity = resp.PServerActivity
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetServerActivity failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

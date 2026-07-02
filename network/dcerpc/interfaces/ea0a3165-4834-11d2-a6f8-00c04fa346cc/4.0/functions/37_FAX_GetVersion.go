package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_GetVersionRequest carries the [in] parameters of FAX_GetVersion.
type fAX_GetVersionRequest struct {
	PVersion msfax.FAX_VERSION
}

func (*fAX_GetVersionRequest) Opnum() uint16 { return fax.OpnumFAX_GetVersion }

// fAX_GetVersionResponse carries the [out] parameters and return value of FAX_GetVersion.
type fAX_GetVersionResponse struct {
	PVersion msfax.FAX_VERSION
	Status   ndr.DWORD `ndr:"retval"`
}

// FAX_GetVersion calls FAX_GetVersion (opnum 37) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetVersion(rpc ndr.Invoker, pVersion msfax.FAX_VERSION) (PVersion msfax.FAX_VERSION, err error) {
	req := &fAX_GetVersionRequest{
		PVersion: pVersion,
	}
	var resp fAX_GetVersionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetVersion: %w", err)
		return
	}
	PVersion = resp.PVersion
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetVersion failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

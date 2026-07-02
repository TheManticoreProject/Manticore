package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_GetConfigOptionRequest carries the [in] parameters of FAX_GetConfigOption.
type fAX_GetConfigOptionRequest struct {
	Option msfax.FAX_ENUM_CONFIG_OPTION
}

func (*fAX_GetConfigOptionRequest) Opnum() uint16 { return fax.OpnumFAX_GetConfigOption }

// fAX_GetConfigOptionResponse carries the [out] parameters and return value of FAX_GetConfigOption.
type fAX_GetConfigOptionResponse struct {
	LpdwValue ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// FAX_GetConfigOption calls FAX_GetConfigOption (opnum 103) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetConfigOption(rpc ndr.Invoker, option msfax.FAX_ENUM_CONFIG_OPTION) (LpdwValue ndr.DWORD, err error) {
	req := &fAX_GetConfigOptionRequest{
		Option: option,
	}
	var resp fAX_GetConfigOptionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetConfigOption: %w", err)
		return
	}
	LpdwValue = resp.LpdwValue
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetConfigOption failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_RegisterServiceProviderExRequest carries the [in] parameters of FAX_RegisterServiceProviderEx.
type fAX_RegisterServiceProviderExRequest struct {
	LpcwstrGUID         ndr.WSTR
	LpcwstrFriendlyName ndr.WSTR
	LpcwstrImageName    ndr.WSTR
	LpcwstrTspName      ndr.WSTR
	DwFSPIVersion       ndr.DWORD
	DwCapabilities      ndr.DWORD
}

func (*fAX_RegisterServiceProviderExRequest) Opnum() uint16 {
	return fax.OpnumFAX_RegisterServiceProviderEx
}

// fAX_RegisterServiceProviderExResponse carries the [out] parameters and return value of FAX_RegisterServiceProviderEx.
type fAX_RegisterServiceProviderExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_RegisterServiceProviderEx calls FAX_RegisterServiceProviderEx (opnum 60) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_RegisterServiceProviderEx(rpc ndr.Invoker, lpcwstrGUID ndr.WSTR, lpcwstrFriendlyName ndr.WSTR, lpcwstrImageName ndr.WSTR, lpcwstrTspName ndr.WSTR, dwFSPIVersion ndr.DWORD, dwCapabilities ndr.DWORD) (err error) {
	req := &fAX_RegisterServiceProviderExRequest{
		LpcwstrGUID:         lpcwstrGUID,
		LpcwstrFriendlyName: lpcwstrFriendlyName,
		LpcwstrImageName:    lpcwstrImageName,
		LpcwstrTspName:      lpcwstrTspName,
		DwFSPIVersion:       dwFSPIVersion,
		DwCapabilities:      dwCapabilities,
	}
	var resp fAX_RegisterServiceProviderExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_RegisterServiceProviderEx: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_RegisterServiceProviderEx failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

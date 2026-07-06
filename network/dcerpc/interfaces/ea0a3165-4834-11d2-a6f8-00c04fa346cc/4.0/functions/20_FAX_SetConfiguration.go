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

// fAX_SetConfigurationRequest carries the [in] parameters of FAX_SetConfiguration.
type fAX_SetConfigurationRequest struct {
	FaxConfig msfax.FAX_CONFIGURATIONW
}

func (*fAX_SetConfigurationRequest) Opnum() uint16 { return fax.OpnumFAX_SetConfiguration }

// fAX_SetConfigurationResponse carries the [out] parameters and return value of FAX_SetConfiguration.
type fAX_SetConfigurationResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetConfiguration calls FAX_SetConfiguration (opnum 20) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetConfiguration(rpc ndr.Invoker, faxConfig msfax.FAX_CONFIGURATIONW) (err error) {
	req := &fAX_SetConfigurationRequest{
		FaxConfig: faxConfig,
	}
	var resp fAX_SetConfigurationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetConfiguration: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetConfiguration failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

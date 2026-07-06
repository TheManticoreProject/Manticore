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

// fAX_SetActivityLoggingConfigurationRequest carries the [in] parameters of FAX_SetActivityLoggingConfiguration.
type fAX_SetActivityLoggingConfigurationRequest struct {
	PActivLogCfg msfax.FAX_ACTIVITY_LOGGING_CONFIGW
}

func (*fAX_SetActivityLoggingConfigurationRequest) Opnum() uint16 {
	return fax.OpnumFAX_SetActivityLoggingConfiguration
}

// fAX_SetActivityLoggingConfigurationResponse carries the [out] parameters and return value of FAX_SetActivityLoggingConfiguration.
type fAX_SetActivityLoggingConfigurationResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetActivityLoggingConfiguration calls FAX_SetActivityLoggingConfiguration (opnum 44) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetActivityLoggingConfiguration(rpc ndr.Invoker, pActivLogCfg msfax.FAX_ACTIVITY_LOGGING_CONFIGW) (err error) {
	req := &fAX_SetActivityLoggingConfigurationRequest{
		PActivLogCfg: pActivLogCfg,
	}
	var resp fAX_SetActivityLoggingConfigurationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetActivityLoggingConfiguration: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetActivityLoggingConfiguration failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

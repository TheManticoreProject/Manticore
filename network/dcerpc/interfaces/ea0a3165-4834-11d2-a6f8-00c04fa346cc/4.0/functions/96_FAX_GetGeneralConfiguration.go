package functions

// IDL source: [MS-FAX] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fax/0907310f-0693-47e7-a6cb-3e599c89a1dd
// A fetched copy is kept at ms-fax.idl in the interface directory.

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_GetGeneralConfigurationRequest carries the [in] parameters of FAX_GetGeneralConfiguration.
type fAX_GetGeneralConfigurationRequest struct {
	Level ndr.DWORD
}

func (*fAX_GetGeneralConfigurationRequest) Opnum() uint16 {
	return fax.OpnumFAX_GetGeneralConfiguration
}

// fAX_GetGeneralConfigurationResponse carries the [out] parameters and return value of FAX_GetGeneralConfiguration.
type fAX_GetGeneralConfigurationResponse struct {
	Buffer     []byte `ndr:"unique,conformant"`
	BufferSize ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// FAX_GetGeneralConfiguration calls FAX_GetGeneralConfiguration (opnum 96) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetGeneralConfiguration(rpc ndr.Invoker, level ndr.DWORD) (Buffer []byte, BufferSize ndr.DWORD, err error) {
	req := &fAX_GetGeneralConfigurationRequest{
		Level: level,
	}
	var resp fAX_GetGeneralConfigurationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetGeneralConfiguration: %w", err)
		return
	}
	Buffer = resp.Buffer
	BufferSize = resp.BufferSize
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetGeneralConfiguration failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

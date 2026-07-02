package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_SetGeneralConfigurationRequest carries the [in] parameters of FAX_SetGeneralConfiguration.
type fAX_SetGeneralConfigurationRequest struct {
	Level      ndr.DWORD
	Buffer     []uint8 `ndr:"ref,size_is=BufferSize"`
	BufferSize ndr.DWORD
}

func (*fAX_SetGeneralConfigurationRequest) Opnum() uint16 {
	return fax.OpnumFAX_SetGeneralConfiguration
}

// fAX_SetGeneralConfigurationResponse carries the [out] parameters and return value of FAX_SetGeneralConfiguration.
type fAX_SetGeneralConfigurationResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetGeneralConfiguration calls FAX_SetGeneralConfiguration (opnum 97) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetGeneralConfiguration(rpc ndr.Invoker, level ndr.DWORD, buffer []uint8, bufferSize ndr.DWORD) (err error) {
	req := &fAX_SetGeneralConfigurationRequest{
		Level:      level,
		Buffer:     buffer,
		BufferSize: bufferSize,
	}
	var resp fAX_SetGeneralConfigurationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetGeneralConfiguration: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetGeneralConfiguration failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

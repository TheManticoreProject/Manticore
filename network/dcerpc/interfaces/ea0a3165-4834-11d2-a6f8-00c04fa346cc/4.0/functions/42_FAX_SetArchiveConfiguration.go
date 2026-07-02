package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_SetArchiveConfigurationRequest carries the [in] parameters of FAX_SetArchiveConfiguration.
type fAX_SetArchiveConfigurationRequest struct {
	Folder      msfax.FAX_ENUM_MESSAGE_FOLDER
	PArchiveCfg uint8
}

func (*fAX_SetArchiveConfigurationRequest) Opnum() uint16 {
	return fax.OpnumFAX_SetArchiveConfiguration
}

// fAX_SetArchiveConfigurationResponse carries the [out] parameters and return value of FAX_SetArchiveConfiguration.
type fAX_SetArchiveConfigurationResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetArchiveConfiguration calls FAX_SetArchiveConfiguration (opnum 42) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetArchiveConfiguration(rpc ndr.Invoker, folder msfax.FAX_ENUM_MESSAGE_FOLDER, pArchiveCfg uint8) (err error) {
	req := &fAX_SetArchiveConfigurationRequest{
		Folder:      folder,
		PArchiveCfg: pArchiveCfg,
	}
	var resp fAX_SetArchiveConfigurationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetArchiveConfiguration: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetArchiveConfiguration failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

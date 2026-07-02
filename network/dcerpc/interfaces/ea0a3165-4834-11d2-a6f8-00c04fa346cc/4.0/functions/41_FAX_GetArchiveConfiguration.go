package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_GetArchiveConfigurationRequest carries the [in] parameters of FAX_GetArchiveConfiguration.
type fAX_GetArchiveConfigurationRequest struct {
	Folder msfax.FAX_ENUM_MESSAGE_FOLDER
}

func (*fAX_GetArchiveConfigurationRequest) Opnum() uint16 {
	return fax.OpnumFAX_GetArchiveConfiguration
}

// fAX_GetArchiveConfigurationResponse carries the [out] parameters and return value of FAX_GetArchiveConfiguration.
type fAX_GetArchiveConfigurationResponse struct {
	Buffer     []byte `ndr:"unique,conformant"`
	BufferSize ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// FAX_GetArchiveConfiguration calls FAX_GetArchiveConfiguration (opnum 41) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetArchiveConfiguration(rpc ndr.Invoker, folder msfax.FAX_ENUM_MESSAGE_FOLDER) (Buffer []byte, BufferSize ndr.DWORD, err error) {
	req := &fAX_GetArchiveConfigurationRequest{
		Folder: folder,
	}
	var resp fAX_GetArchiveConfigurationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetArchiveConfiguration: %w", err)
		return
	}
	Buffer = resp.Buffer
	BufferSize = resp.BufferSize
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetArchiveConfiguration failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

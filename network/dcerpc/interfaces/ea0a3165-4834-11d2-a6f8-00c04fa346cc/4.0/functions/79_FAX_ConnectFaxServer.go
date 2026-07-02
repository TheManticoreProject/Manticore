package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_ConnectFaxServerRequest carries the [in] parameters of FAX_ConnectFaxServer.
type fAX_ConnectFaxServerRequest struct {
	DwClientAPIVersion ndr.DWORD
}

func (*fAX_ConnectFaxServerRequest) Opnum() uint16 { return fax.OpnumFAX_ConnectFaxServer }

// fAX_ConnectFaxServerResponse carries the [out] parameters and return value of FAX_ConnectFaxServer.
type fAX_ConnectFaxServerResponse struct {
	LpdwServerAPIVersion ndr.DWORD
	PHandle              msfax.PRPC_FAX_SVC_HANDLE
	Status               ndr.DWORD `ndr:"retval"`
}

// FAX_ConnectFaxServer calls FAX_ConnectFaxServer (opnum 79) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_ConnectFaxServer(rpc ndr.Invoker, dwClientAPIVersion ndr.DWORD) (LpdwServerAPIVersion ndr.DWORD, PHandle msfax.PRPC_FAX_SVC_HANDLE, err error) {
	req := &fAX_ConnectFaxServerRequest{
		DwClientAPIVersion: dwClientAPIVersion,
	}
	var resp fAX_ConnectFaxServerResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_ConnectFaxServer: %w", err)
		return
	}
	LpdwServerAPIVersion = resp.LpdwServerAPIVersion
	PHandle = resp.PHandle
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_ConnectFaxServer failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

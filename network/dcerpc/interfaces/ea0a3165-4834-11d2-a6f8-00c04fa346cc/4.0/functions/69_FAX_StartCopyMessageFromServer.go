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

// fAX_StartCopyMessageFromServerRequest carries the [in] parameters of FAX_StartCopyMessageFromServer.
type fAX_StartCopyMessageFromServerRequest struct {
	DwlMessageId uint64
	Folder       msfax.FAX_ENUM_MESSAGE_FOLDER
}

func (*fAX_StartCopyMessageFromServerRequest) Opnum() uint16 {
	return fax.OpnumFAX_StartCopyMessageFromServer
}

// fAX_StartCopyMessageFromServerResponse carries the [out] parameters and return value of FAX_StartCopyMessageFromServer.
type fAX_StartCopyMessageFromServerResponse struct {
	LpHandle msfax.PRPC_FAX_COPY_HANDLE
	Status   ndr.DWORD `ndr:"retval"`
}

// FAX_StartCopyMessageFromServer calls FAX_StartCopyMessageFromServer (opnum 69) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_StartCopyMessageFromServer(rpc ndr.Invoker, dwlMessageId uint64, folder msfax.FAX_ENUM_MESSAGE_FOLDER) (LpHandle msfax.PRPC_FAX_COPY_HANDLE, err error) {
	req := &fAX_StartCopyMessageFromServerRequest{
		DwlMessageId: dwlMessageId,
		Folder:       folder,
	}
	var resp fAX_StartCopyMessageFromServerResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_StartCopyMessageFromServer: %w", err)
		return
	}
	LpHandle = resp.LpHandle
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_StartCopyMessageFromServer failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

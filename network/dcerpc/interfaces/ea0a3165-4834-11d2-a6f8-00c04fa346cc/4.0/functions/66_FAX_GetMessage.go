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

// fAX_GetMessageRequest carries the [in] parameters of FAX_GetMessage.
type fAX_GetMessageRequest struct {
	DwlMessageId uint64
	Folder       msfax.FAX_ENUM_MESSAGE_FOLDER
}

func (*fAX_GetMessageRequest) Opnum() uint16 { return fax.OpnumFAX_GetMessage }

// fAX_GetMessageResponse carries the [out] parameters and return value of FAX_GetMessage.
type fAX_GetMessageResponse struct {
	LppBuffer      []byte `ndr:"unique,conformant"`
	LpdwBufferSize ndr.DWORD
	Status         ndr.DWORD `ndr:"retval"`
}

// FAX_GetMessage calls FAX_GetMessage (opnum 66) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetMessage(rpc ndr.Invoker, dwlMessageId uint64, folder msfax.FAX_ENUM_MESSAGE_FOLDER) (LppBuffer []byte, LpdwBufferSize ndr.DWORD, err error) {
	req := &fAX_GetMessageRequest{
		DwlMessageId: dwlMessageId,
		Folder:       folder,
	}
	var resp fAX_GetMessageResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetMessage: %w", err)
		return
	}
	LppBuffer = resp.LppBuffer
	LpdwBufferSize = resp.LpdwBufferSize
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetMessage failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

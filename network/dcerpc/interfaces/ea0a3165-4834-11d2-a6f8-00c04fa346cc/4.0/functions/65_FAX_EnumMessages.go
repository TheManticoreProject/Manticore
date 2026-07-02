package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_EnumMessagesRequest carries the [in] parameters of FAX_EnumMessages.
type fAX_EnumMessagesRequest struct {
	HEnum         msfax.RPC_FAX_MSG_ENUM_HANDLE
	DwNumMessages ndr.DWORD
}

func (*fAX_EnumMessagesRequest) Opnum() uint16 { return fax.OpnumFAX_EnumMessages }

// fAX_EnumMessagesResponse carries the [out] parameters and return value of FAX_EnumMessages.
type fAX_EnumMessagesResponse struct {
	LppBuffer                []byte `ndr:"unique,conformant"`
	LpdwBufferSize           ndr.DWORD
	LpdwNumMessagesRetrieved ndr.DWORD
	Status                   ndr.DWORD `ndr:"retval"`
}

// FAX_EnumMessages calls FAX_EnumMessages (opnum 65) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_EnumMessages(rpc ndr.Invoker, hEnum msfax.RPC_FAX_MSG_ENUM_HANDLE, dwNumMessages ndr.DWORD) (LppBuffer []byte, LpdwBufferSize ndr.DWORD, LpdwNumMessagesRetrieved ndr.DWORD, err error) {
	req := &fAX_EnumMessagesRequest{
		HEnum:         hEnum,
		DwNumMessages: dwNumMessages,
	}
	var resp fAX_EnumMessagesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_EnumMessages: %w", err)
		return
	}
	LppBuffer = resp.LppBuffer
	LpdwBufferSize = resp.LpdwBufferSize
	LpdwNumMessagesRetrieved = resp.LpdwNumMessagesRetrieved
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_EnumMessages failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

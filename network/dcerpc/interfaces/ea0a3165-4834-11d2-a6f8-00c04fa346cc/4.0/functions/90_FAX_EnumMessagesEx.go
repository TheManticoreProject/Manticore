package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_EnumMessagesExRequest carries the [in] parameters of FAX_EnumMessagesEx.
type fAX_EnumMessagesExRequest struct {
	HEnum         msfax.RPC_FAX_MSG_ENUM_HANDLE
	DwNumMessages ndr.DWORD
}

func (*fAX_EnumMessagesExRequest) Opnum() uint16 { return fax.OpnumFAX_EnumMessagesEx }

// fAX_EnumMessagesExResponse carries the [out] parameters and return value of FAX_EnumMessagesEx.
type fAX_EnumMessagesExResponse struct {
	LppBuffer                []byte `ndr:"unique,conformant"`
	LpdwBufferSize           ndr.DWORD
	LpdwNumMessagesRetrieved ndr.DWORD
	LpdwLevel                ndr.DWORD
	Status                   ndr.DWORD `ndr:"retval"`
}

// FAX_EnumMessagesEx calls FAX_EnumMessagesEx (opnum 90) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_EnumMessagesEx(rpc ndr.Invoker, hEnum msfax.RPC_FAX_MSG_ENUM_HANDLE, dwNumMessages ndr.DWORD) (LppBuffer []byte, LpdwBufferSize ndr.DWORD, LpdwNumMessagesRetrieved ndr.DWORD, LpdwLevel ndr.DWORD, err error) {
	req := &fAX_EnumMessagesExRequest{
		HEnum:         hEnum,
		DwNumMessages: dwNumMessages,
	}
	var resp fAX_EnumMessagesExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_EnumMessagesEx: %w", err)
		return
	}
	LppBuffer = resp.LppBuffer
	LpdwBufferSize = resp.LpdwBufferSize
	LpdwNumMessagesRetrieved = resp.LpdwNumMessagesRetrieved
	LpdwLevel = resp.LpdwLevel
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_EnumMessagesEx failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

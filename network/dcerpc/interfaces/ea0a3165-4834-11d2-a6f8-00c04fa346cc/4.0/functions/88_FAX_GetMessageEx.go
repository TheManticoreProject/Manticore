package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_GetMessageExRequest carries the [in] parameters of FAX_GetMessageEx.
type fAX_GetMessageExRequest struct {
	DwlMessageId uint64
	Folder       msfax.FAX_ENUM_MESSAGE_FOLDER
	Level        ndr.DWORD
}

func (*fAX_GetMessageExRequest) Opnum() uint16 { return fax.OpnumFAX_GetMessageEx }

// fAX_GetMessageExResponse carries the [out] parameters and return value of FAX_GetMessageEx.
type fAX_GetMessageExResponse struct {
	LppBuffer      []byte `ndr:"unique,conformant"`
	LpdwBufferSize ndr.DWORD
	Status         ndr.DWORD `ndr:"retval"`
}

// FAX_GetMessageEx calls FAX_GetMessageEx (opnum 88) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetMessageEx(rpc ndr.Invoker, dwlMessageId uint64, folder msfax.FAX_ENUM_MESSAGE_FOLDER, level ndr.DWORD) (LppBuffer []byte, LpdwBufferSize ndr.DWORD, err error) {
	req := &fAX_GetMessageExRequest{
		DwlMessageId: dwlMessageId,
		Folder:       folder,
		Level:        level,
	}
	var resp fAX_GetMessageExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetMessageEx: %w", err)
		return
	}
	LppBuffer = resp.LppBuffer
	LpdwBufferSize = resp.LpdwBufferSize
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetMessageEx failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

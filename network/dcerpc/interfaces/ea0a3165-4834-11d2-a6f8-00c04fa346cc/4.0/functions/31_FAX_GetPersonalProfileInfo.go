package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_GetPersonalProfileInfoRequest carries the [in] parameters of FAX_GetPersonalProfileInfo.
type fAX_GetPersonalProfileInfoRequest struct {
	DwlMessageId uint64
	DwFolder     msfax.FAX_ENUM_MESSAGE_FOLDER
	ProfType     msfax.FAX_ENUM_PERSONAL_PROF_TYPES
}

func (*fAX_GetPersonalProfileInfoRequest) Opnum() uint16 { return fax.OpnumFAX_GetPersonalProfileInfo }

// fAX_GetPersonalProfileInfoResponse carries the [out] parameters and return value of FAX_GetPersonalProfileInfo.
type fAX_GetPersonalProfileInfoResponse struct {
	Buffer     []byte `ndr:"unique,conformant"`
	BufferSize ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// FAX_GetPersonalProfileInfo calls FAX_GetPersonalProfileInfo (opnum 31) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetPersonalProfileInfo(rpc ndr.Invoker, dwlMessageId uint64, dwFolder msfax.FAX_ENUM_MESSAGE_FOLDER, profType msfax.FAX_ENUM_PERSONAL_PROF_TYPES) (Buffer []byte, BufferSize ndr.DWORD, err error) {
	req := &fAX_GetPersonalProfileInfoRequest{
		DwlMessageId: dwlMessageId,
		DwFolder:     dwFolder,
		ProfType:     profType,
	}
	var resp fAX_GetPersonalProfileInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetPersonalProfileInfo: %w", err)
		return
	}
	Buffer = resp.Buffer
	BufferSize = resp.BufferSize
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetPersonalProfileInfo failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

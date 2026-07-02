package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_StartMessagesEnumExRequest carries the [in] parameters of FAX_StartMessagesEnumEx.
type fAX_StartMessagesEnumExRequest struct {
	FAllAccounts       ndr.BOOL
	LpcwstrAccountName *ndr.WSTR `ndr:"unique"`
	Folder             msfax.FAX_ENUM_MESSAGE_FOLDER
	Level              ndr.DWORD
}

func (*fAX_StartMessagesEnumExRequest) Opnum() uint16 { return fax.OpnumFAX_StartMessagesEnumEx }

// fAX_StartMessagesEnumExResponse carries the [out] parameters and return value of FAX_StartMessagesEnumEx.
type fAX_StartMessagesEnumExResponse struct {
	LpHandle msfax.PRPC_FAX_MSG_ENUM_HANDLE
	Status   ndr.DWORD `ndr:"retval"`
}

// FAX_StartMessagesEnumEx calls FAX_StartMessagesEnumEx (opnum 89) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_StartMessagesEnumEx(rpc ndr.Invoker, fAllAccounts ndr.BOOL, lpcwstrAccountName *ndr.WSTR, folder msfax.FAX_ENUM_MESSAGE_FOLDER, level ndr.DWORD) (LpHandle msfax.PRPC_FAX_MSG_ENUM_HANDLE, err error) {
	req := &fAX_StartMessagesEnumExRequest{
		FAllAccounts:       fAllAccounts,
		LpcwstrAccountName: lpcwstrAccountName,
		Folder:             folder,
		Level:              level,
	}
	var resp fAX_StartMessagesEnumExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_StartMessagesEnumEx: %w", err)
		return
	}
	LpHandle = resp.LpHandle
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_StartMessagesEnumEx failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

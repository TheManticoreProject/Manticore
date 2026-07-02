package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_GetAccountInfoRequest carries the [in] parameters of FAX_GetAccountInfo.
type fAX_GetAccountInfoRequest struct {
	LpcwstrAccountName *ndr.WSTR `ndr:"unique"`
	Level              ndr.DWORD
}

func (*fAX_GetAccountInfoRequest) Opnum() uint16 { return fax.OpnumFAX_GetAccountInfo }

// fAX_GetAccountInfoResponse carries the [out] parameters and return value of FAX_GetAccountInfo.
type fAX_GetAccountInfoResponse struct {
	Buffer     []byte `ndr:"unique,conformant"`
	BufferSize ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// FAX_GetAccountInfo calls FAX_GetAccountInfo (opnum 95) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetAccountInfo(rpc ndr.Invoker, lpcwstrAccountName *ndr.WSTR, level ndr.DWORD) (Buffer []byte, BufferSize ndr.DWORD, err error) {
	req := &fAX_GetAccountInfoRequest{
		LpcwstrAccountName: lpcwstrAccountName,
		Level:              level,
	}
	var resp fAX_GetAccountInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetAccountInfo: %w", err)
		return
	}
	Buffer = resp.Buffer
	BufferSize = resp.BufferSize
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetAccountInfo failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

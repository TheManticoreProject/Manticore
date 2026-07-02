package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_EnumAccountsRequest carries the [in] parameters of FAX_EnumAccounts.
type fAX_EnumAccountsRequest struct {
	Level ndr.DWORD
}

func (*fAX_EnumAccountsRequest) Opnum() uint16 { return fax.OpnumFAX_EnumAccounts }

// fAX_EnumAccountsResponse carries the [out] parameters and return value of FAX_EnumAccounts.
type fAX_EnumAccountsResponse struct {
	Buffer       []byte `ndr:"unique,conformant"`
	BufferSize   ndr.DWORD
	LpdwAccounts ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// FAX_EnumAccounts calls FAX_EnumAccounts (opnum 94) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_EnumAccounts(rpc ndr.Invoker, level ndr.DWORD) (Buffer []byte, BufferSize ndr.DWORD, LpdwAccounts ndr.DWORD, err error) {
	req := &fAX_EnumAccountsRequest{
		Level: level,
	}
	var resp fAX_EnumAccountsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_EnumAccounts: %w", err)
		return
	}
	Buffer = resp.Buffer
	BufferSize = resp.BufferSize
	LpdwAccounts = resp.LpdwAccounts
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_EnumAccounts failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

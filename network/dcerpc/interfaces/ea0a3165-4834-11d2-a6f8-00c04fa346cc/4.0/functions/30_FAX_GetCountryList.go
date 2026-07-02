package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_GetCountryListRequest carries the [in] parameters of FAX_GetCountryList.
type fAX_GetCountryListRequest struct {
}

func (*fAX_GetCountryListRequest) Opnum() uint16 { return fax.OpnumFAX_GetCountryList }

// fAX_GetCountryListResponse carries the [out] parameters and return value of FAX_GetCountryList.
type fAX_GetCountryListResponse struct {
	Buffer     []byte `ndr:"unique,conformant"`
	BufferSize ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// FAX_GetCountryList calls FAX_GetCountryList (opnum 30) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetCountryList(rpc ndr.Invoker) (Buffer []byte, BufferSize ndr.DWORD, err error) {
	req := &fAX_GetCountryListRequest{}
	var resp fAX_GetCountryListResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetCountryList: %w", err)
		return
	}
	Buffer = resp.Buffer
	BufferSize = resp.BufferSize
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetCountryList failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

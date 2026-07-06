package functions

// IDL source: [MS-FAX] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fax/0907310f-0693-47e7-a6cb-3e599c89a1dd
// A fetched copy is kept at ms-fax.idl in the interface directory.

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_SetLoggingCategoriesRequest carries the [in] parameters of FAX_SetLoggingCategories.
type fAX_SetLoggingCategoriesRequest struct {
	Buffer           []uint8 `ndr:"ref,size_is=BufferSize"`
	BufferSize       ndr.DWORD
	NumberCategories ndr.DWORD
}

func (*fAX_SetLoggingCategoriesRequest) Opnum() uint16 { return fax.OpnumFAX_SetLoggingCategories }

// fAX_SetLoggingCategoriesResponse carries the [out] parameters and return value of FAX_SetLoggingCategories.
type fAX_SetLoggingCategoriesResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetLoggingCategories calls FAX_SetLoggingCategories (opnum 22) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetLoggingCategories(rpc ndr.Invoker, buffer []uint8, bufferSize ndr.DWORD, numberCategories ndr.DWORD) (err error) {
	req := &fAX_SetLoggingCategoriesRequest{
		Buffer:           buffer,
		BufferSize:       bufferSize,
		NumberCategories: numberCategories,
	}
	var resp fAX_SetLoggingCategoriesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetLoggingCategories: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetLoggingCategories failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

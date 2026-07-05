package functions

import (
	"fmt"

	W32Time "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8fb6d884-2388-11d0-8c35-00c04fda2795/4.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// w32TimeQuerySourceRequest carries the [in] parameters of W32TimeQuerySource.
type w32TimeQuerySourceRequest struct {
}

func (*w32TimeQuerySourceRequest) Opnum() uint16 { return W32Time.OpnumW32TimeQuerySource }

// w32TimeQuerySourceResponse carries the [out] parameters and return value of W32TimeQuerySource.
// pwszSource is [out, string] wchar_t** — the outer pointer is the [ref] out slot and the
// inner [unique,string] pointer is transmitted as a referent id followed by the string body.
type w32TimeQuerySourceResponse struct {
	PwszSource *ndr.WSTR `ndr:"unique"`
	Status     ndr.DWORD `ndr:"retval"`
}

// W32TimeQuerySource calls W32TimeQuerySource (opnum 3) ([MS-W32T] 3.2.4.4). It returns the
// name of the current time source, or nil when the server sends a null string pointer.
func W32TimeQuerySource(rpc ndr.Invoker) (PwszSource *ndr.WSTR, err error) {
	req := &w32TimeQuerySourceRequest{}
	var resp w32TimeQuerySourceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("W32TimeQuerySource: %w", err)
		return
	}
	PwszSource = resp.PwszSource
	if uint32(resp.Status) != W32Time.StatusSuccess {
		err = fmt.Errorf("W32TimeQuerySource failed: %s", W32Time.StatusString(uint32(resp.Status)))
	}
	return
}

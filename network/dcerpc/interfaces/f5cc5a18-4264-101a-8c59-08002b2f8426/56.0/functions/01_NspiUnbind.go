package functions

import (
	"fmt"

	nspi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc5a18-4264-101a-8c59-08002b2f8426/56.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnspi "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nspi"
)

// nspiUnbindRequest carries the [in] parameters of NspiUnbind.
type nspiUnbindRequest struct {
	ContextHandle msnspi.NSPI_HANDLE
	Reserved      ndr.DWORD
}

func (*nspiUnbindRequest) Opnum() uint16 { return nspi.OpnumNspiUnbind }

// nspiUnbindResponse carries the [out] parameters and return value of NspiUnbind.
type nspiUnbindResponse struct {
	ContextHandle msnspi.NSPI_HANDLE
	Status        ndr.DWORD `ndr:"retval"`
}

// NspiUnbind calls NspiUnbind (opnum 1) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiUnbind(rpc ndr.Invoker, contextHandle msnspi.NSPI_HANDLE, reserved ndr.DWORD) (ContextHandle msnspi.NSPI_HANDLE, err error) {
	req := &nspiUnbindRequest{
		ContextHandle: contextHandle,
		Reserved:      reserved,
	}
	var resp nspiUnbindResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiUnbind: %w", err)
		return
	}
	ContextHandle = resp.ContextHandle
	if uint32(resp.Status) != nspi.StatusSuccess {
		err = fmt.Errorf("NspiUnbind failed: %s", nspi.StatusString(uint32(resp.Status)))
	}
	return
}

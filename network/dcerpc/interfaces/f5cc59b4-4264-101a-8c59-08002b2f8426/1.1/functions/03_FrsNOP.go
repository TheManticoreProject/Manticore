package functions

// IDL source: [MS-FRS1] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-frs1/dd60a0d9-176a-46f4-9904-000172041b92
// A fetched copy is kept at ms-frs1.idl in the interface directory.

import (
	"fmt"

	frsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc59b4-4264-101a-8c59-08002b2f8426/1.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// frsNOPRequest carries the [in] parameters of FrsNOP.
type frsNOPRequest struct {
}

func (*frsNOPRequest) Opnum() uint16 { return frsrpc.OpnumFrsNOP }

// FrsNOP calls FrsNOP (opnum 3) ([MS-FRS1] section 3.3.4.3).
func FrsNOP(rpc ndr.Invoker) (err error) {
	req := &frsNOPRequest{}
	var resp statusResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FrsNOP: %w", err)
		return
	}
	if win32.WIN32_ERROR(resp.Status) != win32.ERROR_SUCCESS {
		err = fmt.Errorf("FrsNOP failed: %s", win32.WIN32_ERROR(resp.Status).String())
	}
	return
}

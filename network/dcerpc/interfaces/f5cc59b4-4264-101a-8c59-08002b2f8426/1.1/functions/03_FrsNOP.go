package functions

import (
	"fmt"

	frsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc59b4-4264-101a-8c59-08002b2f8426/1.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
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
	if uint32(resp.Status) != frsrpc.StatusSuccess {
		err = fmt.Errorf("FrsNOP failed: %s", frsrpc.StatusString(uint32(resp.Status)))
	}
	return
}

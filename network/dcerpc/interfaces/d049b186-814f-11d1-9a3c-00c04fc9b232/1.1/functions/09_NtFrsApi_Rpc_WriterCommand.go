package functions

import (
	"fmt"

	frsapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/d049b186-814f-11d1-9a3c-00c04fc9b232/1.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// ntFrsApi_Rpc_WriterCommandRequest carries the [in] parameters of NtFrsApi_Rpc_WriterCommand.
type ntFrsApi_Rpc_WriterCommandRequest struct {
	Command ndr.DWORD
}

func (*ntFrsApi_Rpc_WriterCommandRequest) Opnum() uint16 {
	return frsapi.OpnumNtFrsApi_Rpc_WriterCommand
}

// NtFrsApi_Rpc_WriterCommand calls NtFrsApi_Rpc_WriterCommand (opnum 9) ([MS-FRS1] section 3.2.4.5).
func NtFrsApi_Rpc_WriterCommand(rpc ndr.Invoker, command ndr.DWORD) (err error) {
	req := &ntFrsApi_Rpc_WriterCommandRequest{
		Command: command,
	}
	var resp statusResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NtFrsApi_Rpc_WriterCommand: %w", err)
		return
	}
	if uint32(resp.Status) != frsapi.StatusSuccess {
		err = fmt.Errorf("NtFrsApi_Rpc_WriterCommand failed: %s", frsapi.StatusString(uint32(resp.Status)))
	}
	return
}

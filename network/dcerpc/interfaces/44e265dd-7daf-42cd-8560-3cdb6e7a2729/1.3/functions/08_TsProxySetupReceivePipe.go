package functions

import (
	"fmt"

	tsgu "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/44e265dd-7daf-42cd-8560-3cdb6e7a2729/1.3"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// tsProxySetupReceivePipeRequest carries the [in] parameters of TsProxySetupReceivePipe.
type tsProxySetupReceivePipeRequest struct {
	PRpcMessage []uint8 `ndr:"ref,conformant"`
}

func (*tsProxySetupReceivePipeRequest) Opnum() uint16 {
	return tsgu.OpnumTsProxySetupReceivePipe
}

// tsProxySetupReceivePipeResponse carries the [out] parameters and return value of TsProxySetupReceivePipe.
type tsProxySetupReceivePipeResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// TsProxySetupReceivePipe calls TsProxySetupReceivePipe (opnum 8) ([MS-TSGU] — verify the parameter
// modeling and status handling).
func TsProxySetupReceivePipe(rpc ndr.Invoker, pRpcMessage []uint8) (err error) {
	req := &tsProxySetupReceivePipeRequest{
		PRpcMessage: pRpcMessage,
	}
	var resp tsProxySetupReceivePipeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("TsProxySetupReceivePipe: %w", err)
		return
	}
	if uint32(resp.Status) != tsgu.StatusSuccess {
		err = fmt.Errorf("TsProxySetupReceivePipe failed: %s", tsgu.StatusString(uint32(resp.Status)))
	}
	return
}

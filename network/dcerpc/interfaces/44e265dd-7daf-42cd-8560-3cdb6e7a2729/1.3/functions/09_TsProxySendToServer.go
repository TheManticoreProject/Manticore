package functions

import (
	"fmt"

	tsgu "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/44e265dd-7daf-42cd-8560-3cdb6e7a2729/1.3"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// tsProxySendToServerRequest carries the [in] parameters of TsProxySendToServer.
type tsProxySendToServerRequest struct {
	PRpcMessage []uint8 `ndr:"ref,conformant"`
}

func (*tsProxySendToServerRequest) Opnum() uint16 {
	return tsgu.OpnumTsProxySendToServer
}

// tsProxySendToServerResponse carries the [out] parameters and return value of TsProxySendToServer.
type tsProxySendToServerResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// TsProxySendToServer calls TsProxySendToServer (opnum 9) ([MS-TSGU] — verify the parameter
// modeling and status handling).
func TsProxySendToServer(rpc ndr.Invoker, pRpcMessage []uint8) (err error) {
	req := &tsProxySendToServerRequest{
		PRpcMessage: pRpcMessage,
	}
	var resp tsProxySendToServerResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("TsProxySendToServer: %w", err)
		return
	}
	if uint32(resp.Status) != tsgu.StatusSuccess {
		err = fmt.Errorf("TsProxySendToServer failed: %s", tsgu.StatusString(uint32(resp.Status)))
	}
	return
}

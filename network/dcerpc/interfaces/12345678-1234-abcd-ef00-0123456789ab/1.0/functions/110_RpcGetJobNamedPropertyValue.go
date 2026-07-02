package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcGetJobNamedPropertyValueRequest carries the [in] parameters of RpcGetJobNamedPropertyValue.
type rpcGetJobNamedPropertyValueRequest struct {
	HPrinter structures.PRINTER_HANDLE
	JobId    ndr.DWORD
	PszName  ndr.WSTR
}

func (*rpcGetJobNamedPropertyValueRequest) Opnum() uint16 {
	return winspool.OpnumRpcGetJobNamedPropertyValue
}

// rpcGetJobNamedPropertyValueResponse carries the [out] parameters and return value of RpcGetJobNamedPropertyValue.
type rpcGetJobNamedPropertyValueResponse struct {
	PValue structures.RPC_PrintPropertyValue
	Status ndr.DWORD `ndr:"retval"`
}

// RpcGetJobNamedPropertyValue calls RpcGetJobNamedPropertyValue (opnum 110) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcGetJobNamedPropertyValue(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, jobId ndr.DWORD, pszName ndr.WSTR) (PValue structures.RPC_PrintPropertyValue, err error) {
	req := &rpcGetJobNamedPropertyValueRequest{
		HPrinter: hPrinter,
		JobId:    jobId,
		PszName:  pszName,
	}
	var resp rpcGetJobNamedPropertyValueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetJobNamedPropertyValue: %w", err)
		return
	}
	PValue = resp.PValue
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcGetJobNamedPropertyValue failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}

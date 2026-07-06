package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncDeleteJobNamedPropertyRequest carries the [in] parameters of RpcAsyncDeleteJobNamedProperty.
type rpcAsyncDeleteJobNamedPropertyRequest struct {
	HPrinter mspar.PRINTER_HANDLE
	JobId    ndr.DWORD
	PszName  ndr.WSTR
}

func (*rpcAsyncDeleteJobNamedPropertyRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncDeleteJobNamedProperty
}

// rpcAsyncDeleteJobNamedPropertyResponse carries the [out] parameters and return value of RpcAsyncDeleteJobNamedProperty.
type rpcAsyncDeleteJobNamedPropertyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncDeleteJobNamedProperty calls RpcAsyncDeleteJobNamedProperty (opnum 72) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncDeleteJobNamedProperty(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, jobId ndr.DWORD, pszName ndr.WSTR) (err error) {
	req := &rpcAsyncDeleteJobNamedPropertyRequest{
		HPrinter: hPrinter,
		JobId:    jobId,
		PszName:  pszName,
	}
	var resp rpcAsyncDeleteJobNamedPropertyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncDeleteJobNamedProperty: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncDeleteJobNamedProperty failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}

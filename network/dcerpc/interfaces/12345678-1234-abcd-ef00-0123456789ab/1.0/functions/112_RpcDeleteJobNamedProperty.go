package functions

// IDL source: [MS-RPRN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rprn/e8f9dad8-d114-41cc-9a52-fc927e908cf4
// A fetched copy is kept at ms-rprn.idl in the interface directory.

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcDeleteJobNamedPropertyRequest carries the [in] parameters of RpcDeleteJobNamedProperty.
type rpcDeleteJobNamedPropertyRequest struct {
	HPrinter msrprn.PRINTER_HANDLE
	JobId    ndr.DWORD
	PszName  ndr.WSTR
}

func (*rpcDeleteJobNamedPropertyRequest) Opnum() uint16 {
	return winspool.OpnumRpcDeleteJobNamedProperty
}

// rpcDeleteJobNamedPropertyResponse carries the [out] parameters and return value of RpcDeleteJobNamedProperty.
type rpcDeleteJobNamedPropertyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcDeleteJobNamedProperty calls RpcDeleteJobNamedProperty (opnum 112) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcDeleteJobNamedProperty(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, jobId ndr.DWORD, pszName ndr.WSTR) (err error) {
	req := &rpcDeleteJobNamedPropertyRequest{
		HPrinter: hPrinter,
		JobId:    jobId,
		PszName:  pszName,
	}
	var resp rpcDeleteJobNamedPropertyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcDeleteJobNamedProperty: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcDeleteJobNamedProperty failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}

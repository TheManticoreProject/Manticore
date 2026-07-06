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

// rpcAsyncEnumJobNamedPropertiesRequest carries the [in] parameters of RpcAsyncEnumJobNamedProperties.
type rpcAsyncEnumJobNamedPropertiesRequest struct {
	HPrinter mspar.PRINTER_HANDLE
	JobId    ndr.DWORD
}

func (*rpcAsyncEnumJobNamedPropertiesRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncEnumJobNamedProperties
}

// rpcAsyncEnumJobNamedPropertiesResponse carries the [out] parameters and return value of RpcAsyncEnumJobNamedProperties.
type rpcAsyncEnumJobNamedPropertiesResponse struct {
	PcProperties ndr.DWORD
	PpProperties []mspar.RPC_PrintNamedProperty `ndr:"ref,conformant"`
	Status       ndr.DWORD                      `ndr:"retval"`
}

// RpcAsyncEnumJobNamedProperties calls RpcAsyncEnumJobNamedProperties (opnum 73) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncEnumJobNamedProperties(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, jobId ndr.DWORD) (PcProperties ndr.DWORD, PpProperties []mspar.RPC_PrintNamedProperty, err error) {
	req := &rpcAsyncEnumJobNamedPropertiesRequest{
		HPrinter: hPrinter,
		JobId:    jobId,
	}
	var resp rpcAsyncEnumJobNamedPropertiesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncEnumJobNamedProperties: %w", err)
		return
	}
	PcProperties = resp.PcProperties
	PpProperties = resp.PpProperties
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncEnumJobNamedProperties failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}

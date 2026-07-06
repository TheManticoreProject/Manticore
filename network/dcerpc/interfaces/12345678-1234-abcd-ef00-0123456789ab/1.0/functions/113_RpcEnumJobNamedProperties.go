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

// rpcEnumJobNamedPropertiesRequest carries the [in] parameters of RpcEnumJobNamedProperties.
type rpcEnumJobNamedPropertiesRequest struct {
	HPrinter msrprn.PRINTER_HANDLE
	JobId    ndr.DWORD
}

func (*rpcEnumJobNamedPropertiesRequest) Opnum() uint16 {
	return winspool.OpnumRpcEnumJobNamedProperties
}

// rpcEnumJobNamedPropertiesResponse carries the [out] parameters and return value of RpcEnumJobNamedProperties.
type rpcEnumJobNamedPropertiesResponse struct {
	PcProperties ndr.DWORD
	PpProperties []*msrprn.RPC_PrintNamedProperty `ndr:"elem=unique,ref,conformant"`
	Status       ndr.DWORD                        `ndr:"retval"`
}

// RpcEnumJobNamedProperties calls RpcEnumJobNamedProperties (opnum 113) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcEnumJobNamedProperties(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, jobId ndr.DWORD) (PcProperties ndr.DWORD, PpProperties []*msrprn.RPC_PrintNamedProperty, err error) {
	req := &rpcEnumJobNamedPropertiesRequest{
		HPrinter: hPrinter,
		JobId:    jobId,
	}
	var resp rpcEnumJobNamedPropertiesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcEnumJobNamedProperties: %w", err)
		return
	}
	PcProperties = resp.PcProperties
	PpProperties = resp.PpProperties
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcEnumJobNamedProperties failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}

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

// rpcAsyncLogJobInfoForBranchOfficeRequest carries the [in] parameters of RpcAsyncLogJobInfoForBranchOffice.
type rpcAsyncLogJobInfoForBranchOfficeRequest struct {
	HPrinter                      mspar.PRINTER_HANDLE
	PBranchOfficeJobDataContainer mspar.RPC_BranchOfficeJobDataContainer
}

func (*rpcAsyncLogJobInfoForBranchOfficeRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncLogJobInfoForBranchOffice
}

// rpcAsyncLogJobInfoForBranchOfficeResponse carries the [out] parameters and return value of RpcAsyncLogJobInfoForBranchOffice.
type rpcAsyncLogJobInfoForBranchOfficeResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncLogJobInfoForBranchOffice calls RpcAsyncLogJobInfoForBranchOffice (opnum 74) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncLogJobInfoForBranchOffice(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pBranchOfficeJobDataContainer mspar.RPC_BranchOfficeJobDataContainer) (err error) {
	req := &rpcAsyncLogJobInfoForBranchOfficeRequest{
		HPrinter:                      hPrinter,
		PBranchOfficeJobDataContainer: pBranchOfficeJobDataContainer,
	}
	var resp rpcAsyncLogJobInfoForBranchOfficeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncLogJobInfoForBranchOffice: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncLogJobInfoForBranchOffice failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}

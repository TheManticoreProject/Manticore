package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAsyncEnumPrinterDriversRequest carries the [in] parameters of RpcAsyncEnumPrinterDrivers.
type rpcAsyncEnumPrinterDriversRequest struct {
	PName        *ndr.WSTR `ndr:"unique"`
	PEnvironment *ndr.WSTR `ndr:"unique"`
	Level        ndr.DWORD
	PDrivers     []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf        ndr.DWORD
}

func (*rpcAsyncEnumPrinterDriversRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncEnumPrinterDrivers
}

// rpcAsyncEnumPrinterDriversResponse carries the [out] parameters and return value of RpcAsyncEnumPrinterDrivers.
type rpcAsyncEnumPrinterDriversResponse struct {
	PDrivers   []uint8 `ndr:"ref,size_is=CbBuf"`
	PcbNeeded  ndr.DWORD
	PcReturned ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// RpcAsyncEnumPrinterDrivers calls RpcAsyncEnumPrinterDrivers (opnum 40) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncEnumPrinterDrivers(rpc ndr.Invoker, pName *ndr.WSTR, pEnvironment *ndr.WSTR, level ndr.DWORD, pDrivers []uint8, cbBuf ndr.DWORD) (PDrivers []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcAsyncEnumPrinterDriversRequest{
		PName:        pName,
		PEnvironment: pEnvironment,
		Level:        level,
		PDrivers:     pDrivers,
		CbBuf:        cbBuf,
	}
	var resp rpcAsyncEnumPrinterDriversResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncEnumPrinterDrivers: %w", err)
		return
	}
	PDrivers = resp.PDrivers
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncEnumPrinterDrivers failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}

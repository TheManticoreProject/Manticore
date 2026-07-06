package functions

// IDL source: [MS-RPRN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rprn/e8f9dad8-d114-41cc-9a52-fc927e908cf4
// A fetched copy is kept at ms-rprn.idl in the interface directory.

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcEnumPrinterDriversRequest carries the [in] parameters of RpcEnumPrinterDrivers.
type rpcEnumPrinterDriversRequest struct {
	PName        *ndr.WSTR `ndr:"unique"`
	PEnvironment *ndr.WSTR `ndr:"unique"`
	Level        ndr.DWORD
	PDrivers     []uint8 `ndr:"unique,size_is=CbBuf"`
	CbBuf        ndr.DWORD
}

func (*rpcEnumPrinterDriversRequest) Opnum() uint16 { return winspool.OpnumRpcEnumPrinterDrivers }

// rpcEnumPrinterDriversResponse carries the [out] parameters and return value of RpcEnumPrinterDrivers.
type rpcEnumPrinterDriversResponse struct {
	PDrivers   []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded  ndr.DWORD
	PcReturned ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// RpcEnumPrinterDrivers calls RpcEnumPrinterDrivers (opnum 10) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcEnumPrinterDrivers(rpc ndr.Invoker, pName *ndr.WSTR, pEnvironment *ndr.WSTR, level ndr.DWORD, pDrivers []uint8, cbBuf ndr.DWORD) (PDrivers []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcEnumPrinterDriversRequest{
		PName:        pName,
		PEnvironment: pEnvironment,
		Level:        level,
		PDrivers:     pDrivers,
		CbBuf:        cbBuf,
	}
	var resp rpcEnumPrinterDriversResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcEnumPrinterDrivers: %w", err)
		return
	}
	PDrivers = resp.PDrivers
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcEnumPrinterDrivers failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}

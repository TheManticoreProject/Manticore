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

// rpcEnumPrintersRequest carries the [in] parameters of RpcEnumPrinters.
type rpcEnumPrintersRequest struct {
	Flags        ndr.DWORD
	Name         *ndr.WSTR `ndr:"unique"`
	Level        ndr.DWORD
	PPrinterEnum []uint8 `ndr:"unique,size_is=CbBuf"`
	CbBuf        ndr.DWORD
}

func (*rpcEnumPrintersRequest) Opnum() uint16 { return winspool.OpnumRpcEnumPrinters }

// rpcEnumPrintersResponse carries the [out] parameters and return value of RpcEnumPrinters.
type rpcEnumPrintersResponse struct {
	PPrinterEnum []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded    ndr.DWORD
	PcReturned   ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// RpcEnumPrinters calls RpcEnumPrinters (opnum 0) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcEnumPrinters(rpc ndr.Invoker, flags ndr.DWORD, name *ndr.WSTR, level ndr.DWORD, pPrinterEnum []uint8, cbBuf ndr.DWORD) (PPrinterEnum []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcEnumPrintersRequest{
		Flags:        flags,
		Name:         name,
		Level:        level,
		PPrinterEnum: pPrinterEnum,
		CbBuf:        cbBuf,
	}
	var resp rpcEnumPrintersResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcEnumPrinters: %w", err)
		return
	}
	PPrinterEnum = resp.PPrinterEnum
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcEnumPrinters failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}

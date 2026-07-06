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

// rpcEnumPrinterDataExRequest carries the [in] parameters of RpcEnumPrinterDataEx.
type rpcEnumPrinterDataExRequest struct {
	HPrinter     msrprn.PRINTER_HANDLE
	PKeyName     ndr.WSTR
	CbEnumValues ndr.DWORD
}

func (*rpcEnumPrinterDataExRequest) Opnum() uint16 { return winspool.OpnumRpcEnumPrinterDataEx }

// rpcEnumPrinterDataExResponse carries the [out] parameters and return value of RpcEnumPrinterDataEx.
type rpcEnumPrinterDataExResponse struct {
	PEnumValues   []uint8 `ndr:"ref,size_is=CbEnumValues"`
	PcbEnumValues ndr.DWORD
	PnEnumValues  ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// RpcEnumPrinterDataEx calls RpcEnumPrinterDataEx (opnum 79) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcEnumPrinterDataEx(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, pKeyName ndr.WSTR, cbEnumValues ndr.DWORD) (PEnumValues []uint8, PcbEnumValues ndr.DWORD, PnEnumValues ndr.DWORD, err error) {
	req := &rpcEnumPrinterDataExRequest{
		HPrinter:     hPrinter,
		PKeyName:     pKeyName,
		CbEnumValues: cbEnumValues,
	}
	var resp rpcEnumPrinterDataExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcEnumPrinterDataEx: %w", err)
		return
	}
	PEnumValues = resp.PEnumValues
	PcbEnumValues = resp.PcbEnumValues
	PnEnumValues = resp.PnEnumValues
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcEnumPrinterDataEx failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}

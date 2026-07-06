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

// rpcAsyncEnumPrinterKeyRequest carries the [in] parameters of RpcAsyncEnumPrinterKey.
type rpcAsyncEnumPrinterKeyRequest struct {
	HPrinter mspar.PRINTER_HANDLE
	PKeyName ndr.WSTR
	CbSubkey ndr.DWORD
}

func (*rpcAsyncEnumPrinterKeyRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncEnumPrinterKey
}

// rpcAsyncEnumPrinterKeyResponse carries the [out] parameters and return value of RpcAsyncEnumPrinterKey.
type rpcAsyncEnumPrinterKeyResponse struct {
	PSubkey   []uint16 `ndr:"ref,size_is=CbSubkey"`
	PcbSubkey ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcAsyncEnumPrinterKey calls RpcAsyncEnumPrinterKey (opnum 29) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncEnumPrinterKey(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pKeyName ndr.WSTR, cbSubkey ndr.DWORD) (PSubkey []uint16, PcbSubkey ndr.DWORD, err error) {
	req := &rpcAsyncEnumPrinterKeyRequest{
		HPrinter: hPrinter,
		PKeyName: pKeyName,
		CbSubkey: cbSubkey,
	}
	var resp rpcAsyncEnumPrinterKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncEnumPrinterKey: %w", err)
		return
	}
	PSubkey = resp.PSubkey
	PcbSubkey = resp.PcbSubkey
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncEnumPrinterKey failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}

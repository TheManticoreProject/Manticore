package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncPlayGdiScriptOnPrinterICRequest carries the [in] parameters of RpcAsyncPlayGdiScriptOnPrinterIC.
type rpcAsyncPlayGdiScriptOnPrinterICRequest struct {
	HPrinterIC mspar.GDI_HANDLE
	PIn        []uint8 `ndr:"ref,size_is=CIn"`
	CIn        ndr.DWORD
	COut       ndr.DWORD
	Ul         ndr.DWORD
}

func (*rpcAsyncPlayGdiScriptOnPrinterICRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncPlayGdiScriptOnPrinterIC
}

// rpcAsyncPlayGdiScriptOnPrinterICResponse carries the [out] parameters and return value of RpcAsyncPlayGdiScriptOnPrinterIC.
type rpcAsyncPlayGdiScriptOnPrinterICResponse struct {
	POut   []uint8   `ndr:"ref,size_is=COut"`
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncPlayGdiScriptOnPrinterIC calls RpcAsyncPlayGdiScriptOnPrinterIC (opnum 36) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncPlayGdiScriptOnPrinterIC(rpc ndr.Invoker, hPrinterIC mspar.GDI_HANDLE, pIn []uint8, cIn ndr.DWORD, cOut ndr.DWORD, ul ndr.DWORD) (POut []uint8, err error) {
	req := &rpcAsyncPlayGdiScriptOnPrinterICRequest{
		HPrinterIC: hPrinterIC,
		PIn:        pIn,
		CIn:        cIn,
		COut:       cOut,
		Ul:         ul,
	}
	var resp rpcAsyncPlayGdiScriptOnPrinterICResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncPlayGdiScriptOnPrinterIC: %w", err)
		return
	}
	POut = resp.POut
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncPlayGdiScriptOnPrinterIC failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}

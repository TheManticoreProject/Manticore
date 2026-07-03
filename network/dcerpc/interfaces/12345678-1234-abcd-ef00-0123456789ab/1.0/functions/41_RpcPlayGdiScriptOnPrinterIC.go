package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcPlayGdiScriptOnPrinterICRequest carries the [in] parameters of RpcPlayGdiScriptOnPrinterIC.
type rpcPlayGdiScriptOnPrinterICRequest struct {
	HPrinterIC msrprn.GDI_HANDLE
	PIn        []uint8 `ndr:"ref,size_is=CIn"`
	CIn        ndr.DWORD
	COut       ndr.DWORD
	Ul         ndr.DWORD
}

func (*rpcPlayGdiScriptOnPrinterICRequest) Opnum() uint16 {
	return winspool.OpnumRpcPlayGdiScriptOnPrinterIC
}

// rpcPlayGdiScriptOnPrinterICResponse carries the [out] parameters and return value of RpcPlayGdiScriptOnPrinterIC.
type rpcPlayGdiScriptOnPrinterICResponse struct {
	POut   []uint8   `ndr:"ref,size_is=COut"`
	Status ndr.DWORD `ndr:"retval"`
}

// RpcPlayGdiScriptOnPrinterIC calls RpcPlayGdiScriptOnPrinterIC (opnum 41) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcPlayGdiScriptOnPrinterIC(rpc ndr.Invoker, hPrinterIC msrprn.GDI_HANDLE, pIn []uint8, cIn ndr.DWORD, cOut ndr.DWORD, ul ndr.DWORD) (POut []uint8, err error) {
	req := &rpcPlayGdiScriptOnPrinterICRequest{
		HPrinterIC: hPrinterIC,
		PIn:        pIn,
		CIn:        cIn,
		COut:       cOut,
		Ul:         ul,
	}
	var resp rpcPlayGdiScriptOnPrinterICResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcPlayGdiScriptOnPrinterIC: %w", err)
		return
	}
	POut = resp.POut
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcPlayGdiScriptOnPrinterIC failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}

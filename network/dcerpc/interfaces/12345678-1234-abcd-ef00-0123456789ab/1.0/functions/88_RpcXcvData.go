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

// rpcXcvDataRequest carries the [in] parameters of RpcXcvData.
type rpcXcvDataRequest struct {
	HXcv         msrprn.PRINTER_HANDLE
	PszDataName  ndr.WSTR
	PInputData   []uint8 `ndr:"ref,size_is=CbInputData"`
	CbInputData  ndr.DWORD
	CbOutputData ndr.DWORD
	PdwStatus    ndr.DWORD
}

func (*rpcXcvDataRequest) Opnum() uint16 { return winspool.OpnumRpcXcvData }

// rpcXcvDataResponse carries the [out] parameters and return value of RpcXcvData.
type rpcXcvDataResponse struct {
	POutputData     []uint8 `ndr:"ref,size_is=CbOutputData"`
	PcbOutputNeeded ndr.DWORD
	PdwStatus       ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// RpcXcvData calls RpcXcvData (opnum 88) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcXcvData(rpc ndr.Invoker, hXcv msrprn.PRINTER_HANDLE, pszDataName ndr.WSTR, pInputData []uint8, cbInputData ndr.DWORD, cbOutputData ndr.DWORD, pdwStatus ndr.DWORD) (POutputData []uint8, PcbOutputNeeded ndr.DWORD, PdwStatus ndr.DWORD, err error) {
	req := &rpcXcvDataRequest{
		HXcv:         hXcv,
		PszDataName:  pszDataName,
		PInputData:   pInputData,
		CbInputData:  cbInputData,
		CbOutputData: cbOutputData,
		PdwStatus:    pdwStatus,
	}
	var resp rpcXcvDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcXcvData: %w", err)
		return
	}
	POutputData = resp.POutputData
	PcbOutputNeeded = resp.PcbOutputNeeded
	PdwStatus = resp.PdwStatus
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcXcvData failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}

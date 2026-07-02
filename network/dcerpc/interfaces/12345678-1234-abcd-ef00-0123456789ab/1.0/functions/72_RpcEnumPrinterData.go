package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcEnumPrinterDataRequest carries the [in] parameters of RpcEnumPrinterData.
type rpcEnumPrinterDataRequest struct {
	HPrinter    structures.PRINTER_HANDLE
	DwIndex     ndr.DWORD
	CbValueName ndr.DWORD
	CbData      ndr.DWORD
}

func (*rpcEnumPrinterDataRequest) Opnum() uint16 { return winspool.OpnumRpcEnumPrinterData }

// rpcEnumPrinterDataResponse carries the [out] parameters and return value of RpcEnumPrinterData.
type rpcEnumPrinterDataResponse struct {
	PValueName   []uint16 `ndr:"ref,conformant"`
	PcbValueName ndr.DWORD
	PType        ndr.DWORD
	PData        []uint8 `ndr:"ref,size_is=CbData"`
	PcbData      ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// RpcEnumPrinterData calls RpcEnumPrinterData (opnum 72) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcEnumPrinterData(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, dwIndex ndr.DWORD, cbValueName ndr.DWORD, cbData ndr.DWORD) (PValueName []uint16, PcbValueName ndr.DWORD, PType ndr.DWORD, PData []uint8, PcbData ndr.DWORD, err error) {
	req := &rpcEnumPrinterDataRequest{
		HPrinter:    hPrinter,
		DwIndex:     dwIndex,
		CbValueName: cbValueName,
		CbData:      cbData,
	}
	var resp rpcEnumPrinterDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcEnumPrinterData: %w", err)
		return
	}
	PValueName = resp.PValueName
	PcbValueName = resp.PcbValueName
	PType = resp.PType
	PData = resp.PData
	PcbData = resp.PcbData
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcEnumPrinterData failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	nspi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc5a18-4264-101a-8c59-08002b2f8426/56.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnspi "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nspi"
)

// nspiGetSpecialTableRequest carries the [in] parameters of NspiGetSpecialTable.
type nspiGetSpecialTableRequest struct {
	HRpc      msnspi.NSPI_HANDLE
	DwFlags   ndr.DWORD
	PStat     msnspi.STAT
	LpVersion ndr.DWORD
}

func (*nspiGetSpecialTableRequest) Opnum() uint16 { return nspi.OpnumNspiGetSpecialTable }

// nspiGetSpecialTableResponse carries the [out] parameters and return value of NspiGetSpecialTable.
type nspiGetSpecialTableResponse struct {
	LpVersion ndr.DWORD
	PpRows    *msnspi.PropertyRowSet_r `ndr:"unique"`
	Status    ndr.DWORD                `ndr:"retval"`
}

// NspiGetSpecialTable calls NspiGetSpecialTable (opnum 12) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiGetSpecialTable(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, dwFlags ndr.DWORD, pStat msnspi.STAT, lpVersion ndr.DWORD) (LpVersion ndr.DWORD, PpRows *msnspi.PropertyRowSet_r, err error) {
	req := &nspiGetSpecialTableRequest{
		HRpc:      hRpc,
		DwFlags:   dwFlags,
		PStat:     pStat,
		LpVersion: lpVersion,
	}
	var resp nspiGetSpecialTableResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiGetSpecialTable: %w", err)
		return
	}
	LpVersion = resp.LpVersion
	PpRows = resp.PpRows
	if uint32(resp.Status) != nspi.StatusSuccess {
		err = fmt.Errorf("NspiGetSpecialTable failed: %s", nspi.StatusString(uint32(resp.Status)))
	}
	return
}

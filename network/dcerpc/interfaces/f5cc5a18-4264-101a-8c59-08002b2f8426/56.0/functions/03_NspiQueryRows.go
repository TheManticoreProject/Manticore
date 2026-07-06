package functions

// IDL source: [MS-NSPI] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nspi/2554418c-a060-473a-950a-e009a53e33d9
// A fetched copy is kept at ms-nspi.idl in the interface directory.

import (
	"fmt"

	nspi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc5a18-4264-101a-8c59-08002b2f8426/56.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnspi "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nspi"
)

// nspiQueryRowsRequest carries the [in] parameters of NspiQueryRows.
type nspiQueryRowsRequest struct {
	HRpc          msnspi.NSPI_HANDLE
	DwFlags       ndr.DWORD
	PStat         msnspi.STAT
	DwETableCount ndr.DWORD
	LpETable      []ndr.DWORD `ndr:"unique,size_is=DwETableCount"`
	Count         ndr.DWORD
	PPropTags     *msnspi.PropertyTagArray_r `ndr:"unique"`
}

func (*nspiQueryRowsRequest) Opnum() uint16 { return nspi.OpnumNspiQueryRows }

// nspiQueryRowsResponse carries the [out] parameters and return value of NspiQueryRows.
type nspiQueryRowsResponse struct {
	PStat  msnspi.STAT
	PpRows *msnspi.PropertyRowSet_r `ndr:"unique"`
	Status ndr.DWORD                `ndr:"retval"`
}

// NspiQueryRows calls NspiQueryRows (opnum 3) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiQueryRows(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, dwFlags ndr.DWORD, pStat msnspi.STAT, dwETableCount ndr.DWORD, lpETable []ndr.DWORD, count ndr.DWORD, pPropTags *msnspi.PropertyTagArray_r) (PStat msnspi.STAT, PpRows *msnspi.PropertyRowSet_r, err error) {
	req := &nspiQueryRowsRequest{
		HRpc:          hRpc,
		DwFlags:       dwFlags,
		PStat:         pStat,
		DwETableCount: dwETableCount,
		LpETable:      lpETable,
		Count:         count,
		PPropTags:     pPropTags,
	}
	var resp nspiQueryRowsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiQueryRows: %w", err)
		return
	}
	PStat = resp.PStat
	PpRows = resp.PpRows
	// ErrorsReturned is a success-severity warning: the call succeeded overall but one or
	// more returned properties are PtypErrorCode ([MS-NSPI] 3.1.4.1.10 / [MS-OXCDATA] 2.5).
	if s := uint32(resp.Status); s != nspi.StatusSuccess && s != nspi.StatusErrorsReturned {
		err = fmt.Errorf("NspiQueryRows failed: %s", nspi.StatusString(s))
	}
	return
}

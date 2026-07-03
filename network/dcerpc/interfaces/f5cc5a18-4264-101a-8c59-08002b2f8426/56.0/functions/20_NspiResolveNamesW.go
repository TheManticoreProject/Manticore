package functions

import (
	"fmt"

	nspi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc5a18-4264-101a-8c59-08002b2f8426/56.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnspi "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nspi"
)

// nspiResolveNamesWRequest carries the [in] parameters of NspiResolveNamesW.
type nspiResolveNamesWRequest struct {
	HRpc      msnspi.NSPI_HANDLE
	Reserved  ndr.DWORD
	PStat     msnspi.STAT
	PPropTags *msnspi.PropertyTagArray_r `ndr:"unique"`
	PaWStr    msnspi.WStringsArray_r
}

func (*nspiResolveNamesWRequest) Opnum() uint16 { return nspi.OpnumNspiResolveNamesW }

// nspiResolveNamesWResponse carries the [out] parameters and return value of NspiResolveNamesW.
type nspiResolveNamesWResponse struct {
	PpMIds *msnspi.PropertyTagArray_r `ndr:"unique"`
	PpRows *msnspi.PropertyRowSet_r   `ndr:"unique"`
	Status ndr.DWORD                  `ndr:"retval"`
}

// NspiResolveNamesW calls NspiResolveNamesW (opnum 20) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiResolveNamesW(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, reserved ndr.DWORD, pStat msnspi.STAT, pPropTags *msnspi.PropertyTagArray_r, paWStr msnspi.WStringsArray_r) (PpMIds *msnspi.PropertyTagArray_r, PpRows *msnspi.PropertyRowSet_r, err error) {
	req := &nspiResolveNamesWRequest{
		HRpc:      hRpc,
		Reserved:  reserved,
		PStat:     pStat,
		PPropTags: pPropTags,
		PaWStr:    paWStr,
	}
	var resp nspiResolveNamesWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiResolveNamesW: %w", err)
		return
	}
	PpMIds = resp.PpMIds
	PpRows = resp.PpRows
	// ErrorsReturned is a success-severity warning: the call succeeded overall but one or
	// more returned properties are PtypErrorCode ([MS-NSPI] 3.1.4.1.20 / [MS-OXCDATA] 2.5).
	if s := uint32(resp.Status); s != nspi.StatusSuccess && s != nspi.StatusErrorsReturned {
		err = fmt.Errorf("NspiResolveNamesW failed: %s", nspi.StatusString(s))
	}
	return
}

package functions

import (
	"fmt"

	nspi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc5a18-4264-101a-8c59-08002b2f8426/56.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnspi "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nspi"
)

// nspiSeekEntriesRequest carries the [in] parameters of NspiSeekEntries.
type nspiSeekEntriesRequest struct {
	HRpc      msnspi.NSPI_HANDLE
	Reserved  ndr.DWORD
	PStat     msnspi.STAT
	PTarget   msnspi.PropertyValue_r
	LpETable  *msnspi.PropertyTagArray_r `ndr:"unique"`
	PPropTags *msnspi.PropertyTagArray_r `ndr:"unique"`
}

func (*nspiSeekEntriesRequest) Opnum() uint16 { return nspi.OpnumNspiSeekEntries }

// nspiSeekEntriesResponse carries the [out] parameters and return value of NspiSeekEntries.
type nspiSeekEntriesResponse struct {
	PStat  msnspi.STAT
	PpRows *msnspi.PropertyRowSet_r `ndr:"unique"`
	Status ndr.DWORD                `ndr:"retval"`
}

// NspiSeekEntries calls NspiSeekEntries (opnum 4) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiSeekEntries(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, reserved ndr.DWORD, pStat msnspi.STAT, pTarget msnspi.PropertyValue_r, lpETable *msnspi.PropertyTagArray_r, pPropTags *msnspi.PropertyTagArray_r) (PStat msnspi.STAT, PpRows *msnspi.PropertyRowSet_r, err error) {
	req := &nspiSeekEntriesRequest{
		HRpc:      hRpc,
		Reserved:  reserved,
		PStat:     pStat,
		PTarget:   pTarget,
		LpETable:  lpETable,
		PPropTags: pPropTags,
	}
	// pTarget's PROP_VAL_UNION discriminant is (ulPropTag & 0xFFFF), transmitted inline;
	// derive it so the selected arm marshals correctly.
	req.PTarget.SetDiscriminant()
	var resp nspiSeekEntriesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiSeekEntries: %w", err)
		return
	}
	PStat = resp.PStat
	PpRows = resp.PpRows
	// ErrorsReturned is a success-severity warning: the call succeeded overall but one or
	// more returned properties are PtypErrorCode ([MS-NSPI] 3.1.4.1.9 / [MS-OXCDATA] 2.5).
	if s := uint32(resp.Status); s != nspi.StatusSuccess && s != nspi.StatusErrorsReturned {
		err = fmt.Errorf("NspiSeekEntries failed: %s", nspi.StatusString(s))
	}
	return
}

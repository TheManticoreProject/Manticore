package functions

import (
	"fmt"

	nspi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc5a18-4264-101a-8c59-08002b2f8426/56.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnspi "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nspi"
)

// nspiGetMatchesRequest carries the [in] parameters of NspiGetMatches.
type nspiGetMatchesRequest struct {
	HRpc        msnspi.NSPI_HANDLE
	Reserved1   ndr.DWORD
	PStat       msnspi.STAT
	PReserved   *msnspi.PropertyTagArray_r `ndr:"unique"`
	Reserved2   ndr.DWORD
	Filter      *msnspi.Restriction_r  `ndr:"unique"`
	LpPropName  *msnspi.PropertyName_r `ndr:"unique"`
	UlRequested ndr.DWORD
	PPropTags   *msnspi.PropertyTagArray_r `ndr:"unique"`
}

func (*nspiGetMatchesRequest) Opnum() uint16 { return nspi.OpnumNspiGetMatches }

// nspiGetMatchesResponse carries the [out] parameters and return value of NspiGetMatches.
type nspiGetMatchesResponse struct {
	PStat     msnspi.STAT
	PpOutMIds *msnspi.PropertyTagArray_r `ndr:"unique"`
	PpRows    *msnspi.PropertyRowSet_r   `ndr:"unique"`
	Status    ndr.DWORD                  `ndr:"retval"`
}

// NspiGetMatches calls NspiGetMatches (opnum 5) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiGetMatches(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, reserved1 ndr.DWORD, pStat msnspi.STAT, pReserved *msnspi.PropertyTagArray_r, reserved2 ndr.DWORD, filter *msnspi.Restriction_r, lpPropName *msnspi.PropertyName_r, ulRequested ndr.DWORD, pPropTags *msnspi.PropertyTagArray_r) (PStat msnspi.STAT, PpOutMIds *msnspi.PropertyTagArray_r, PpRows *msnspi.PropertyRowSet_r, err error) {
	req := &nspiGetMatchesRequest{
		HRpc:        hRpc,
		Reserved1:   reserved1,
		PStat:       pStat,
		PReserved:   pReserved,
		Reserved2:   reserved2,
		Filter:      filter,
		LpPropName:  lpPropName,
		UlRequested: ulRequested,
		PPropTags:   pPropTags,
	}
	// The filter is a non-encapsulated Restriction_r union tree whose discriminants are
	// transmitted inline; derive them from rt/ulPropTag so every arm marshals correctly.
	if req.Filter != nil {
		req.Filter.SetDiscriminants()
	}
	var resp nspiGetMatchesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiGetMatches: %w", err)
		return
	}
	PStat = resp.PStat
	PpOutMIds = resp.PpOutMIds
	PpRows = resp.PpRows
	// ErrorsReturned is a success-severity warning: the call succeeded overall but one or
	// more returned properties are PtypErrorCode ([MS-NSPI] 3.1.4.1.11 / [MS-OXCDATA] 2.5).
	if s := uint32(resp.Status); s != nspi.StatusSuccess && s != nspi.StatusErrorsReturned {
		err = fmt.Errorf("NspiGetMatches failed: %s", nspi.StatusString(s))
	}
	return
}

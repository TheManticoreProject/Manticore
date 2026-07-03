package functions

import (
	"fmt"

	nspi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc5a18-4264-101a-8c59-08002b2f8426/56.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnspi "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nspi"
)

// nspiGetNamesFromIDsRequest carries the [in] parameters of NspiGetNamesFromIDs.
type nspiGetNamesFromIDsRequest struct {
	HRpc      msnspi.NSPI_HANDLE
	Reserved  ndr.DWORD
	Lpguid    *msnspi.FlatUID_r          `ndr:"unique"`
	PPropTags *msnspi.PropertyTagArray_r `ndr:"unique"`
}

func (*nspiGetNamesFromIDsRequest) Opnum() uint16 { return nspi.OpnumNspiGetNamesFromIDs }

// nspiGetNamesFromIDsResponse carries the [out] parameters and return value of NspiGetNamesFromIDs.
type nspiGetNamesFromIDsResponse struct {
	PpReturnedPropTags *msnspi.PropertyTagArray_r `ndr:"unique"`
	PpNames            *msnspi.PropertyNameSet_r  `ndr:"unique"`
	Status             ndr.DWORD                  `ndr:"retval"`
}

// NspiGetNamesFromIDs calls NspiGetNamesFromIDs (opnum 17) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiGetNamesFromIDs(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, reserved ndr.DWORD, lpguid *msnspi.FlatUID_r, pPropTags *msnspi.PropertyTagArray_r) (PpReturnedPropTags *msnspi.PropertyTagArray_r, PpNames *msnspi.PropertyNameSet_r, err error) {
	req := &nspiGetNamesFromIDsRequest{
		HRpc:      hRpc,
		Reserved:  reserved,
		Lpguid:    lpguid,
		PPropTags: pPropTags,
	}
	var resp nspiGetNamesFromIDsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiGetNamesFromIDs: %w", err)
		return
	}
	PpReturnedPropTags = resp.PpReturnedPropTags
	PpNames = resp.PpNames
	if uint32(resp.Status) != nspi.StatusSuccess {
		err = fmt.Errorf("NspiGetNamesFromIDs failed: %s", nspi.StatusString(uint32(resp.Status)))
	}
	return
}

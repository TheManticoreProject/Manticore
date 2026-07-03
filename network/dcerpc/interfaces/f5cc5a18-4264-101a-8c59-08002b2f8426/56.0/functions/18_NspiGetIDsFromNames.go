package functions

import (
	"fmt"

	nspi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc5a18-4264-101a-8c59-08002b2f8426/56.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnspi "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nspi"
)

// nspiGetIDsFromNamesRequest carries the [in] parameters of NspiGetIDsFromNames.
type nspiGetIDsFromNamesRequest struct {
	HRpc       msnspi.NSPI_HANDLE
	Reserved   ndr.DWORD
	DwFlags    ndr.DWORD
	CPropNames ndr.DWORD
	PNames     []*msnspi.PropertyName_r `ndr:"ref,size_is=CPropNames,elem=unique"`
}

func (*nspiGetIDsFromNamesRequest) Opnum() uint16 { return nspi.OpnumNspiGetIDsFromNames }

// nspiGetIDsFromNamesResponse carries the [out] parameters and return value of NspiGetIDsFromNames.
type nspiGetIDsFromNamesResponse struct {
	PpPropTags *msnspi.PropertyTagArray_r `ndr:"unique"`
	Status     ndr.DWORD                  `ndr:"retval"`
}

// NspiGetIDsFromNames calls NspiGetIDsFromNames (opnum 18) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiGetIDsFromNames(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, reserved ndr.DWORD, dwFlags ndr.DWORD, cPropNames ndr.DWORD, pNames []*msnspi.PropertyName_r) (PpPropTags *msnspi.PropertyTagArray_r, err error) {
	req := &nspiGetIDsFromNamesRequest{
		HRpc:       hRpc,
		Reserved:   reserved,
		DwFlags:    dwFlags,
		CPropNames: cPropNames,
		PNames:     pNames,
	}
	var resp nspiGetIDsFromNamesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiGetIDsFromNames: %w", err)
		return
	}
	PpPropTags = resp.PpPropTags
	if uint32(resp.Status) != nspi.StatusSuccess {
		err = fmt.Errorf("NspiGetIDsFromNames failed: %s", nspi.StatusString(uint32(resp.Status)))
	}
	return
}

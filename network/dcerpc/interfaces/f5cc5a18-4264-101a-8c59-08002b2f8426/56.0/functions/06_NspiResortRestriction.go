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

// nspiResortRestrictionRequest carries the [in] parameters of NspiResortRestriction.
type nspiResortRestrictionRequest struct {
	HRpc      msnspi.NSPI_HANDLE
	Reserved  ndr.DWORD
	PStat     msnspi.STAT
	PInMIds   msnspi.PropertyTagArray_r
	PpOutMIds *msnspi.PropertyTagArray_r `ndr:"unique"`
}

func (*nspiResortRestrictionRequest) Opnum() uint16 { return nspi.OpnumNspiResortRestriction }

// nspiResortRestrictionResponse carries the [out] parameters and return value of NspiResortRestriction.
type nspiResortRestrictionResponse struct {
	PStat     msnspi.STAT
	PpOutMIds *msnspi.PropertyTagArray_r `ndr:"unique"`
	Status    ndr.DWORD                  `ndr:"retval"`
}

// NspiResortRestriction calls NspiResortRestriction (opnum 6) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiResortRestriction(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, reserved ndr.DWORD, pStat msnspi.STAT, pInMIds msnspi.PropertyTagArray_r, ppOutMIds *msnspi.PropertyTagArray_r) (PStat msnspi.STAT, PpOutMIds *msnspi.PropertyTagArray_r, err error) {
	req := &nspiResortRestrictionRequest{
		HRpc:      hRpc,
		Reserved:  reserved,
		PStat:     pStat,
		PInMIds:   pInMIds,
		PpOutMIds: ppOutMIds,
	}
	var resp nspiResortRestrictionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiResortRestriction: %w", err)
		return
	}
	PStat = resp.PStat
	PpOutMIds = resp.PpOutMIds
	if uint32(resp.Status) != nspi.StatusSuccess {
		err = fmt.Errorf("NspiResortRestriction failed: %s", nspi.StatusString(uint32(resp.Status)))
	}
	return
}

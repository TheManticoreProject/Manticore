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

// nspiGetPropListRequest carries the [in] parameters of NspiGetPropList.
type nspiGetPropListRequest struct {
	HRpc     msnspi.NSPI_HANDLE
	DwFlags  ndr.DWORD
	DwMId    ndr.DWORD
	CodePage ndr.DWORD
}

func (*nspiGetPropListRequest) Opnum() uint16 { return nspi.OpnumNspiGetPropList }

// nspiGetPropListResponse carries the [out] parameters and return value of NspiGetPropList.
type nspiGetPropListResponse struct {
	PpPropTags *msnspi.PropertyTagArray_r `ndr:"unique"`
	Status     ndr.DWORD                  `ndr:"retval"`
}

// NspiGetPropList calls NspiGetPropList (opnum 8) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiGetPropList(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, dwFlags ndr.DWORD, dwMId ndr.DWORD, codePage ndr.DWORD) (PpPropTags *msnspi.PropertyTagArray_r, err error) {
	req := &nspiGetPropListRequest{
		HRpc:     hRpc,
		DwFlags:  dwFlags,
		DwMId:    dwMId,
		CodePage: codePage,
	}
	var resp nspiGetPropListResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiGetPropList: %w", err)
		return
	}
	PpPropTags = resp.PpPropTags
	if uint32(resp.Status) != nspi.StatusSuccess {
		err = fmt.Errorf("NspiGetPropList failed: %s", nspi.StatusString(uint32(resp.Status)))
	}
	return
}

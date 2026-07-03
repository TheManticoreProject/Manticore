package functions

import (
	"fmt"

	nspi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc5a18-4264-101a-8c59-08002b2f8426/56.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnspi "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nspi"
)

// nspiGetTemplateInfoRequest carries the [in] parameters of NspiGetTemplateInfo.
type nspiGetTemplateInfoRequest struct {
	HRpc       msnspi.NSPI_HANDLE
	DwFlags    ndr.DWORD
	UlType     ndr.DWORD
	PDN        *ndr.STR `ndr:"unique"`
	DwCodePage ndr.DWORD
	DwLocaleID ndr.DWORD
}

func (*nspiGetTemplateInfoRequest) Opnum() uint16 { return nspi.OpnumNspiGetTemplateInfo }

// nspiGetTemplateInfoResponse carries the [out] parameters and return value of NspiGetTemplateInfo.
type nspiGetTemplateInfoResponse struct {
	PpData *msnspi.PropertyRow_r `ndr:"unique"`
	Status ndr.DWORD             `ndr:"retval"`
}

// NspiGetTemplateInfo calls NspiGetTemplateInfo (opnum 13) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiGetTemplateInfo(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, dwFlags ndr.DWORD, ulType ndr.DWORD, pDN *ndr.STR, dwCodePage ndr.DWORD, dwLocaleID ndr.DWORD) (PpData *msnspi.PropertyRow_r, err error) {
	req := &nspiGetTemplateInfoRequest{
		HRpc:       hRpc,
		DwFlags:    dwFlags,
		UlType:     ulType,
		PDN:        pDN,
		DwCodePage: dwCodePage,
		DwLocaleID: dwLocaleID,
	}
	var resp nspiGetTemplateInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiGetTemplateInfo: %w", err)
		return
	}
	PpData = resp.PpData
	// ErrorsReturned is a success-severity warning: the call succeeded overall but one or
	// more returned properties are PtypErrorCode ([MS-NSPI] 3.1.4.1.14 / [MS-OXCDATA] 2.5).
	if s := uint32(resp.Status); s != nspi.StatusSuccess && s != nspi.StatusErrorsReturned {
		err = fmt.Errorf("NspiGetTemplateInfo failed: %s", nspi.StatusString(s))
	}
	return
}

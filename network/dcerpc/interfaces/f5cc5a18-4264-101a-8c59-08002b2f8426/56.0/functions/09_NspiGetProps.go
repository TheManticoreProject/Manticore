package functions

import (
	"fmt"

	nspi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc5a18-4264-101a-8c59-08002b2f8426/56.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnspi "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nspi"
)

// nspiGetPropsRequest carries the [in] parameters of NspiGetProps.
type nspiGetPropsRequest struct {
	HRpc      msnspi.NSPI_HANDLE
	DwFlags   ndr.DWORD
	PStat     msnspi.STAT
	PPropTags *msnspi.PropertyTagArray_r `ndr:"unique"`
}

func (*nspiGetPropsRequest) Opnum() uint16 { return nspi.OpnumNspiGetProps }

// nspiGetPropsResponse carries the [out] parameters and return value of NspiGetProps.
type nspiGetPropsResponse struct {
	PpRows *msnspi.PropertyRow_r `ndr:"unique"`
	Status ndr.DWORD             `ndr:"retval"`
}

// NspiGetProps calls NspiGetProps (opnum 9) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiGetProps(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, dwFlags ndr.DWORD, pStat msnspi.STAT, pPropTags *msnspi.PropertyTagArray_r) (PpRows *msnspi.PropertyRow_r, err error) {
	req := &nspiGetPropsRequest{
		HRpc:      hRpc,
		DwFlags:   dwFlags,
		PStat:     pStat,
		PPropTags: pPropTags,
	}
	var resp nspiGetPropsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiGetProps: %w", err)
		return
	}
	PpRows = resp.PpRows
	// ErrorsReturned is a success-severity warning: the call succeeded overall but one or
	// more returned properties are PtypErrorCode ([MS-NSPI] 3.1.4.1.7 / [MS-OXCDATA] 2.5).
	if s := uint32(resp.Status); s != nspi.StatusSuccess && s != nspi.StatusErrorsReturned {
		err = fmt.Errorf("NspiGetProps failed: %s", nspi.StatusString(s))
	}
	return
}

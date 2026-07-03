package functions

import (
	"fmt"

	nspi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc5a18-4264-101a-8c59-08002b2f8426/56.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnspi "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nspi"
)

// nspiModPropsRequest carries the [in] parameters of NspiModProps.
type nspiModPropsRequest struct {
	HRpc      msnspi.NSPI_HANDLE
	Reserved  ndr.DWORD
	PStat     msnspi.STAT
	PPropTags *msnspi.PropertyTagArray_r `ndr:"unique"`
	PRow      msnspi.PropertyRow_r
}

func (*nspiModPropsRequest) Opnum() uint16 { return nspi.OpnumNspiModProps }

// nspiModPropsResponse carries the [out] parameters and return value of NspiModProps.
type nspiModPropsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NspiModProps calls NspiModProps (opnum 11) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiModProps(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, reserved ndr.DWORD, pStat msnspi.STAT, pPropTags *msnspi.PropertyTagArray_r, pRow msnspi.PropertyRow_r) (err error) {
	req := &nspiModPropsRequest{
		HRpc:      hRpc,
		Reserved:  reserved,
		PStat:     pStat,
		PPropTags: pPropTags,
		PRow:      pRow,
	}
	// The rows carry non-encapsulated PROP_VAL_UNION values whose discriminant is derived
	// from each ulPropTag and transmitted inline; set it so the arms marshal correctly.
	req.PRow.SetDiscriminants()
	var resp nspiModPropsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiModProps: %w", err)
		return
	}
	if uint32(resp.Status) != nspi.StatusSuccess {
		err = fmt.Errorf("NspiModProps failed: %s", nspi.StatusString(uint32(resp.Status)))
	}
	return
}

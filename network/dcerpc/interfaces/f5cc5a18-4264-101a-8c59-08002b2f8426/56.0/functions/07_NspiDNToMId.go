package functions

import (
	"fmt"

	nspi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc5a18-4264-101a-8c59-08002b2f8426/56.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnspi "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nspi"
)

// nspiDNToMIdRequest carries the [in] parameters of NspiDNToMId.
type nspiDNToMIdRequest struct {
	HRpc     msnspi.NSPI_HANDLE
	Reserved ndr.DWORD
	PNames   msnspi.StringsArray_r
}

func (*nspiDNToMIdRequest) Opnum() uint16 { return nspi.OpnumNspiDNToMId }

// nspiDNToMIdResponse carries the [out] parameters and return value of NspiDNToMId.
type nspiDNToMIdResponse struct {
	PpOutMIds *msnspi.PropertyTagArray_r `ndr:"unique"`
	Status    ndr.DWORD                  `ndr:"retval"`
}

// NspiDNToMId calls NspiDNToMId (opnum 7) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiDNToMId(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, reserved ndr.DWORD, pNames msnspi.StringsArray_r) (PpOutMIds *msnspi.PropertyTagArray_r, err error) {
	req := &nspiDNToMIdRequest{
		HRpc:     hRpc,
		Reserved: reserved,
		PNames:   pNames,
	}
	var resp nspiDNToMIdResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiDNToMId: %w", err)
		return
	}
	PpOutMIds = resp.PpOutMIds
	if uint32(resp.Status) != nspi.StatusSuccess {
		err = fmt.Errorf("NspiDNToMId failed: %s", nspi.StatusString(uint32(resp.Status)))
	}
	return
}

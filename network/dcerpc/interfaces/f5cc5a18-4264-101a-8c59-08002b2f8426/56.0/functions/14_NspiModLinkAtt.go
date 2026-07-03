package functions

import (
	"fmt"

	nspi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc5a18-4264-101a-8c59-08002b2f8426/56.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnspi "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nspi"
)

// nspiModLinkAttRequest carries the [in] parameters of NspiModLinkAtt.
type nspiModLinkAttRequest struct {
	HRpc       msnspi.NSPI_HANDLE
	DwFlags    ndr.DWORD
	UlPropTag  ndr.DWORD
	DwMId      ndr.DWORD
	LpEntryIds msnspi.BinaryArray_r
}

func (*nspiModLinkAttRequest) Opnum() uint16 { return nspi.OpnumNspiModLinkAtt }

// nspiModLinkAttResponse carries the [out] parameters and return value of NspiModLinkAtt.
type nspiModLinkAttResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NspiModLinkAtt calls NspiModLinkAtt (opnum 14) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiModLinkAtt(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, dwFlags ndr.DWORD, ulPropTag ndr.DWORD, dwMId ndr.DWORD, lpEntryIds msnspi.BinaryArray_r) (err error) {
	req := &nspiModLinkAttRequest{
		HRpc:       hRpc,
		DwFlags:    dwFlags,
		UlPropTag:  ulPropTag,
		DwMId:      dwMId,
		LpEntryIds: lpEntryIds,
	}
	var resp nspiModLinkAttResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiModLinkAtt: %w", err)
		return
	}
	if uint32(resp.Status) != nspi.StatusSuccess {
		err = fmt.Errorf("NspiModLinkAtt failed: %s", nspi.StatusString(uint32(resp.Status)))
	}
	return
}

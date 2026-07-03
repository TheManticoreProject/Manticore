package functions

import (
	"fmt"

	nspi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc5a18-4264-101a-8c59-08002b2f8426/56.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnspi "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nspi"
)

// nspiCompareMIdsRequest carries the [in] parameters of NspiCompareMIds.
type nspiCompareMIdsRequest struct {
	HRpc     msnspi.NSPI_HANDLE
	Reserved ndr.DWORD
	PStat    msnspi.STAT
	MId1     ndr.DWORD
	MId2     ndr.DWORD
}

func (*nspiCompareMIdsRequest) Opnum() uint16 { return nspi.OpnumNspiCompareMIds }

// nspiCompareMIdsResponse carries the [out] parameters and return value of NspiCompareMIds.
type nspiCompareMIdsResponse struct {
	PlResult int32
	Status   ndr.DWORD `ndr:"retval"`
}

// NspiCompareMIds calls NspiCompareMIds (opnum 10) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiCompareMIds(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, reserved ndr.DWORD, pStat msnspi.STAT, mId1 ndr.DWORD, mId2 ndr.DWORD) (PlResult int32, err error) {
	req := &nspiCompareMIdsRequest{
		HRpc:     hRpc,
		Reserved: reserved,
		PStat:    pStat,
		MId1:     mId1,
		MId2:     mId2,
	}
	var resp nspiCompareMIdsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiCompareMIds: %w", err)
		return
	}
	PlResult = resp.PlResult
	if uint32(resp.Status) != nspi.StatusSuccess {
		err = fmt.Errorf("NspiCompareMIds failed: %s", nspi.StatusString(uint32(resp.Status)))
	}
	return
}

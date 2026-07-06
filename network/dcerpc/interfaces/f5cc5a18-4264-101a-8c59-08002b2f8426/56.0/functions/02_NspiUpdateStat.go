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

// nspiUpdateStatRequest carries the [in] parameters of NspiUpdateStat.
type nspiUpdateStatRequest struct {
	HRpc     msnspi.NSPI_HANDLE
	Reserved ndr.DWORD
	PStat    msnspi.STAT
	PlDelta  *int32 `ndr:"unique"`
}

func (*nspiUpdateStatRequest) Opnum() uint16 { return nspi.OpnumNspiUpdateStat }

// nspiUpdateStatResponse carries the [out] parameters and return value of NspiUpdateStat.
type nspiUpdateStatResponse struct {
	PStat   msnspi.STAT
	PlDelta *int32    `ndr:"unique"`
	Status  ndr.DWORD `ndr:"retval"`
}

// NspiUpdateStat calls NspiUpdateStat (opnum 2) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiUpdateStat(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, reserved ndr.DWORD, pStat msnspi.STAT, plDelta *int32) (PStat msnspi.STAT, PlDelta *int32, err error) {
	req := &nspiUpdateStatRequest{
		HRpc:     hRpc,
		Reserved: reserved,
		PStat:    pStat,
		PlDelta:  plDelta,
	}
	var resp nspiUpdateStatResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiUpdateStat: %w", err)
		return
	}
	PStat = resp.PStat
	PlDelta = resp.PlDelta
	if uint32(resp.Status) != nspi.StatusSuccess {
		err = fmt.Errorf("NspiUpdateStat failed: %s", nspi.StatusString(uint32(resp.Status)))
	}
	return
}

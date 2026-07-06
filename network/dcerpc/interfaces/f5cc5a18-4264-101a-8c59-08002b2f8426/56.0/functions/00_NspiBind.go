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

// nspiBindRequest carries the [in] parameters of NspiBind.
type nspiBindRequest struct {
	DwFlags     ndr.DWORD
	PStat       msnspi.STAT
	PServerGuid *msnspi.FlatUID_r `ndr:"unique"`
}

func (*nspiBindRequest) Opnum() uint16 { return nspi.OpnumNspiBind }

// nspiBindResponse carries the [out] parameters and return value of NspiBind.
type nspiBindResponse struct {
	PServerGuid   *msnspi.FlatUID_r `ndr:"unique"`
	ContextHandle msnspi.NSPI_HANDLE
	Status        ndr.DWORD `ndr:"retval"`
}

// NspiBind calls NspiBind (opnum 0) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiBind(rpc ndr.Invoker, dwFlags ndr.DWORD, pStat msnspi.STAT, pServerGuid *msnspi.FlatUID_r) (PServerGuid *msnspi.FlatUID_r, ContextHandle msnspi.NSPI_HANDLE, err error) {
	req := &nspiBindRequest{
		DwFlags:     dwFlags,
		PStat:       pStat,
		PServerGuid: pServerGuid,
	}
	var resp nspiBindResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiBind: %w", err)
		return
	}
	PServerGuid = resp.PServerGuid
	ContextHandle = resp.ContextHandle
	if uint32(resp.Status) != nspi.StatusSuccess {
		err = fmt.Errorf("NspiBind failed: %s", nspi.StatusString(uint32(resp.Status)))
	}
	return
}

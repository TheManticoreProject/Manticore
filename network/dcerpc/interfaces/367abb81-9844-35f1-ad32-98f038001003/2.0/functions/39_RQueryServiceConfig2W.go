package functions

// IDL source: [MS-SCMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-scmr/19168537-40b5-4d7a-99e0-d77f0f5e0241
// A fetched copy is kept at ms-scmr.idl in the interface directory.

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
)

// rQueryServiceConfig2WRequest carries the [in] parameters of RQueryServiceConfig2W.
type rQueryServiceConfig2WRequest struct {
	HService    msscmr.SC_RPC_HANDLE
	DwInfoLevel ndr.DWORD
	CbBufSize   ndr.DWORD
}

func (*rQueryServiceConfig2WRequest) Opnum() uint16 { return svcctl.OpnumRQueryServiceConfig2W }

// rQueryServiceConfig2WResponse carries the [out] parameters and return value of RQueryServiceConfig2W.
type rQueryServiceConfig2WResponse struct {
	LpBuffer       []uint8 `ndr:"ref,size_is=CbBufSize"`
	PcbBytesNeeded msscmr.LPBOUNDED_DWORD_8K
	Status         ndr.DWORD `ndr:"retval"`
}

// RQueryServiceConfig2W calls RQueryServiceConfig2W (opnum 39) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RQueryServiceConfig2W(rpc ndr.Invoker, hService msscmr.SC_RPC_HANDLE, dwInfoLevel ndr.DWORD, cbBufSize ndr.DWORD) (LpBuffer []uint8, PcbBytesNeeded msscmr.LPBOUNDED_DWORD_8K, err error) {
	req := &rQueryServiceConfig2WRequest{
		HService:    hService,
		DwInfoLevel: dwInfoLevel,
		CbBufSize:   cbBufSize,
	}
	var resp rQueryServiceConfig2WResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RQueryServiceConfig2W: %w", err)
		return
	}
	LpBuffer = resp.LpBuffer
	PcbBytesNeeded = resp.PcbBytesNeeded
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RQueryServiceConfig2W failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}

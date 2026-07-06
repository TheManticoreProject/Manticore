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

// rQueryServiceConfig2ARequest carries the [in] parameters of RQueryServiceConfig2A.
type rQueryServiceConfig2ARequest struct {
	HService    msscmr.SC_RPC_HANDLE
	DwInfoLevel ndr.DWORD
	CbBufSize   ndr.DWORD
}

func (*rQueryServiceConfig2ARequest) Opnum() uint16 { return svcctl.OpnumRQueryServiceConfig2A }

// rQueryServiceConfig2AResponse carries the [out] parameters and return value of RQueryServiceConfig2A.
type rQueryServiceConfig2AResponse struct {
	LpBuffer       []uint8 `ndr:"ref,size_is=CbBufSize"`
	PcbBytesNeeded msscmr.LPBOUNDED_DWORD_8K
	Status         ndr.DWORD `ndr:"retval"`
}

// RQueryServiceConfig2A calls RQueryServiceConfig2A (opnum 38) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RQueryServiceConfig2A(rpc ndr.Invoker, hService msscmr.SC_RPC_HANDLE, dwInfoLevel ndr.DWORD, cbBufSize ndr.DWORD) (LpBuffer []uint8, PcbBytesNeeded msscmr.LPBOUNDED_DWORD_8K, err error) {
	req := &rQueryServiceConfig2ARequest{
		HService:    hService,
		DwInfoLevel: dwInfoLevel,
		CbBufSize:   cbBufSize,
	}
	var resp rQueryServiceConfig2AResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RQueryServiceConfig2A: %w", err)
		return
	}
	LpBuffer = resp.LpBuffer
	PcbBytesNeeded = resp.PcbBytesNeeded
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RQueryServiceConfig2A failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}

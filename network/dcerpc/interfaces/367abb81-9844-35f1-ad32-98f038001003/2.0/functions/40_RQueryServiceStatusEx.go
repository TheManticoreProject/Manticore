package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
)

// rQueryServiceStatusExRequest carries the [in] parameters of RQueryServiceStatusEx.
type rQueryServiceStatusExRequest struct {
	HService  msscmr.SC_RPC_HANDLE
	InfoLevel msscmr.SC_STATUS_TYPE
	CbBufSize ndr.DWORD
}

func (*rQueryServiceStatusExRequest) Opnum() uint16 { return svcctl.OpnumRQueryServiceStatusEx }

// rQueryServiceStatusExResponse carries the [out] parameters and return value of RQueryServiceStatusEx.
type rQueryServiceStatusExResponse struct {
	LpBuffer       []uint8 `ndr:"ref,size_is=CbBufSize"`
	PcbBytesNeeded msscmr.LPBOUNDED_DWORD_8K
	Status         ndr.DWORD `ndr:"retval"`
}

// RQueryServiceStatusEx calls RQueryServiceStatusEx (opnum 40) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RQueryServiceStatusEx(rpc ndr.Invoker, hService msscmr.SC_RPC_HANDLE, infoLevel msscmr.SC_STATUS_TYPE, cbBufSize ndr.DWORD) (LpBuffer []uint8, PcbBytesNeeded msscmr.LPBOUNDED_DWORD_8K, err error) {
	req := &rQueryServiceStatusExRequest{
		HService:  hService,
		InfoLevel: infoLevel,
		CbBufSize: cbBufSize,
	}
	var resp rQueryServiceStatusExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RQueryServiceStatusEx: %w", err)
		return
	}
	LpBuffer = resp.LpBuffer
	PcbBytesNeeded = resp.PcbBytesNeeded
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RQueryServiceStatusEx failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}

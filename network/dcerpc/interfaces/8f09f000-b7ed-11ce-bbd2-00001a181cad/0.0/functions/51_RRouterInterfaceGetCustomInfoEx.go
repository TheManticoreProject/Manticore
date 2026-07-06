package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRouterInterfaceGetCustomInfoExRequest carries the [in] parameters of RRouterInterfaceGetCustomInfoEx.
type rRouterInterfaceGetCustomInfoExRequest struct {
	HInterface      ndr.DWORD
	PIfCustomConfig msrrasm.MPR_IF_CUSTOMINFOEX_IDL
}

func (*rRouterInterfaceGetCustomInfoExRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceGetCustomInfoEx
}

// rRouterInterfaceGetCustomInfoExResponse carries the [out] parameters and return value of RRouterInterfaceGetCustomInfoEx.
type rRouterInterfaceGetCustomInfoExResponse struct {
	PIfCustomConfig msrrasm.MPR_IF_CUSTOMINFOEX_IDL
	Status          ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceGetCustomInfoEx calls RRouterInterfaceGetCustomInfoEx (opnum 51) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceGetCustomInfoEx(rpc ndr.Invoker, hInterface ndr.DWORD, pIfCustomConfig msrrasm.MPR_IF_CUSTOMINFOEX_IDL) (PIfCustomConfig msrrasm.MPR_IF_CUSTOMINFOEX_IDL, err error) {
	req := &rRouterInterfaceGetCustomInfoExRequest{
		HInterface:      hInterface,
		PIfCustomConfig: pIfCustomConfig,
	}
	var resp rRouterInterfaceGetCustomInfoExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceGetCustomInfoEx: %w", err)
		return
	}
	PIfCustomConfig = resp.PIfCustomConfig
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceGetCustomInfoEx failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}

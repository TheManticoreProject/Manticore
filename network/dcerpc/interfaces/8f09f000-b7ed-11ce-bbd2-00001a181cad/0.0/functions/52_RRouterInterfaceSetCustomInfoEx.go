package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRouterInterfaceSetCustomInfoExRequest carries the [in] parameters of RRouterInterfaceSetCustomInfoEx.
type rRouterInterfaceSetCustomInfoExRequest struct {
	HInterface      ndr.DWORD
	PIfCustomConfig msrrasm.MPR_IF_CUSTOMINFOEX_IDL
}

func (*rRouterInterfaceSetCustomInfoExRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceSetCustomInfoEx
}

// rRouterInterfaceSetCustomInfoExResponse carries the [out] parameters and return value of RRouterInterfaceSetCustomInfoEx.
type rRouterInterfaceSetCustomInfoExResponse struct {
	PIfCustomConfig msrrasm.MPR_IF_CUSTOMINFOEX_IDL
	Status          ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceSetCustomInfoEx calls RRouterInterfaceSetCustomInfoEx (opnum 52) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceSetCustomInfoEx(rpc ndr.Invoker, hInterface ndr.DWORD, pIfCustomConfig msrrasm.MPR_IF_CUSTOMINFOEX_IDL) (PIfCustomConfig msrrasm.MPR_IF_CUSTOMINFOEX_IDL, err error) {
	req := &rRouterInterfaceSetCustomInfoExRequest{
		HInterface:      hInterface,
		PIfCustomConfig: pIfCustomConfig,
	}
	var resp rRouterInterfaceSetCustomInfoExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceSetCustomInfoEx: %w", err)
		return
	}
	PIfCustomConfig = resp.PIfCustomConfig
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceSetCustomInfoEx failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rControlServiceExWRequest carries the [in] parameters of RControlServiceExW.
type rControlServiceExWRequest struct {
	HService         structures.SC_RPC_HANDLE
	DwControl        ndr.DWORD
	DwInfoLevel      ndr.DWORD
	PControlInParams structures.SC_RPC_SERVICE_CONTROL_IN_PARAMSW
}

func (*rControlServiceExWRequest) Opnum() uint16 { return svcctl.OpnumRControlServiceExW }

// rControlServiceExWResponse carries the [out] parameters and return value of RControlServiceExW.
type rControlServiceExWResponse struct {
	PControlOutParams structures.SC_RPC_SERVICE_CONTROL_OUT_PARAMSW
	Status            ndr.DWORD `ndr:"retval"`
}

// RControlServiceExW calls RControlServiceExW (opnum 51) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RControlServiceExW(rpc ndr.Invoker, hService structures.SC_RPC_HANDLE, dwControl ndr.DWORD, dwInfoLevel ndr.DWORD, pControlInParams structures.SC_RPC_SERVICE_CONTROL_IN_PARAMSW) (PControlOutParams structures.SC_RPC_SERVICE_CONTROL_OUT_PARAMSW, err error) {
	// The non-encapsulated union re-transmits its discriminant inline ([C706] 14.3.8), so
	// keep it in step with the dwInfoLevel argument that selects the arm.
	pControlInParams.Tag = dwInfoLevel
	req := &rControlServiceExWRequest{
		HService:         hService,
		DwControl:        dwControl,
		DwInfoLevel:      dwInfoLevel,
		PControlInParams: pControlInParams,
	}
	var resp rControlServiceExWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RControlServiceExW: %w", err)
		return
	}
	PControlOutParams = resp.PControlOutParams
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RControlServiceExW failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}

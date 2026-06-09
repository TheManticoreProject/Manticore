package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rControlServiceExARequest carries the [in] parameters of RControlServiceExA.
type rControlServiceExARequest struct {
	HService         structures.SC_RPC_HANDLE
	DwControl        ndr.DWORD
	DwInfoLevel      ndr.DWORD
	PControlInParams structures.SC_RPC_SERVICE_CONTROL_IN_PARAMSA
}

func (*rControlServiceExARequest) Opnum() uint16 { return svcctl.OpnumRControlServiceExA }

// rControlServiceExAResponse carries the [out] parameters and return value of RControlServiceExA.
type rControlServiceExAResponse struct {
	PControlOutParams structures.SC_RPC_SERVICE_CONTROL_OUT_PARAMSA
	Status            ndr.DWORD `ndr:"retval"`
}

// RControlServiceExA calls RControlServiceExA (opnum 50) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RControlServiceExA(rpc ndr.Invoker, hService structures.SC_RPC_HANDLE, dwControl ndr.DWORD, dwInfoLevel ndr.DWORD, pControlInParams structures.SC_RPC_SERVICE_CONTROL_IN_PARAMSA) (PControlOutParams structures.SC_RPC_SERVICE_CONTROL_OUT_PARAMSA, err error) {
	// The non-encapsulated union re-transmits its discriminant inline ([C706] 14.3.8), so
	// keep it in step with the dwInfoLevel argument that selects the arm.
	pControlInParams.Tag = dwInfoLevel
	req := &rControlServiceExARequest{
		HService:         hService,
		DwControl:        dwControl,
		DwInfoLevel:      dwInfoLevel,
		PControlInParams: pControlInParams,
	}
	var resp rControlServiceExAResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RControlServiceExA: %w", err)
		return
	}
	PControlOutParams = resp.PControlOutParams
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RControlServiceExA failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}

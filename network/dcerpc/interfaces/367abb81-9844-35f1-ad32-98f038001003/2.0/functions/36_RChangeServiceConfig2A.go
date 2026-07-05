package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
)

// rChangeServiceConfig2ARequest carries the [in] parameters of RChangeServiceConfig2A.
type rChangeServiceConfig2ARequest struct {
	HService msscmr.SC_RPC_HANDLE
	Info     msscmr.SC_RPC_CONFIG_INFOA
}

func (*rChangeServiceConfig2ARequest) Opnum() uint16 { return svcctl.OpnumRChangeServiceConfig2A }

// rChangeServiceConfig2AResponse carries the [out] parameters and return value of RChangeServiceConfig2A.
type rChangeServiceConfig2AResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RChangeServiceConfig2A calls RChangeServiceConfig2A (opnum 36) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RChangeServiceConfig2A(rpc ndr.Invoker, hService msscmr.SC_RPC_HANDLE, info msscmr.SC_RPC_CONFIG_INFOA) (err error) {
	// dwInfoLevel is re-transmitted inline as the union discriminant ([C706] 14.3.8); keep
	// the two in step so the caller only has to set Info.DwInfoLevel and the matching arm.
	info.Field.Tag = info.DwInfoLevel
	req := &rChangeServiceConfig2ARequest{
		HService: hService,
		Info:     info,
	}
	var resp rChangeServiceConfig2AResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RChangeServiceConfig2A: %w", err)
		return
	}
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RChangeServiceConfig2A failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}

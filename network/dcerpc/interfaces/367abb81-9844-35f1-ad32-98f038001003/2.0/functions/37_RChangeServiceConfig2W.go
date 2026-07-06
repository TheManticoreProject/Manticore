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

// rChangeServiceConfig2WRequest carries the [in] parameters of RChangeServiceConfig2W.
type rChangeServiceConfig2WRequest struct {
	HService msscmr.SC_RPC_HANDLE
	Info     msscmr.SC_RPC_CONFIG_INFOW
}

func (*rChangeServiceConfig2WRequest) Opnum() uint16 { return svcctl.OpnumRChangeServiceConfig2W }

// rChangeServiceConfig2WResponse carries the [out] parameters and return value of RChangeServiceConfig2W.
type rChangeServiceConfig2WResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RChangeServiceConfig2W calls RChangeServiceConfig2W (opnum 37) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RChangeServiceConfig2W(rpc ndr.Invoker, hService msscmr.SC_RPC_HANDLE, info msscmr.SC_RPC_CONFIG_INFOW) (err error) {
	// dwInfoLevel is re-transmitted inline as the union discriminant ([C706] 14.3.8); keep
	// the two in step so the caller only has to set Info.DwInfoLevel and the matching arm.
	info.Field.Tag = info.DwInfoLevel
	req := &rChangeServiceConfig2WRequest{
		HService: hService,
		Info:     info,
	}
	var resp rChangeServiceConfig2WResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RChangeServiceConfig2W: %w", err)
		return
	}
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RChangeServiceConfig2W failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}

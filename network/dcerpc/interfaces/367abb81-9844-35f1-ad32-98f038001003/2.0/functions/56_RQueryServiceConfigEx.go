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

// rQueryServiceConfigExRequest carries the [in] parameters of RQueryServiceConfigEx.
type rQueryServiceConfigExRequest struct {
	HService    msscmr.SC_RPC_HANDLE
	DwInfoLevel ndr.DWORD
}

func (*rQueryServiceConfigExRequest) Opnum() uint16 { return svcctl.OpnumRQueryServiceConfigEx }

// rQueryServiceConfigExResponse carries the [out] parameters and return value of RQueryServiceConfigEx.
type rQueryServiceConfigExResponse struct {
	PInfo  msscmr.SC_RPC_CONFIG_INFOW
	Status ndr.DWORD `ndr:"retval"`
}

// RQueryServiceConfigEx calls RQueryServiceConfigEx (opnum 56) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RQueryServiceConfigEx(rpc ndr.Invoker, hService msscmr.SC_RPC_HANDLE, dwInfoLevel ndr.DWORD) (PInfo msscmr.SC_RPC_CONFIG_INFOW, err error) {
	req := &rQueryServiceConfigExRequest{
		HService:    hService,
		DwInfoLevel: dwInfoLevel,
	}
	var resp rQueryServiceConfigExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RQueryServiceConfigEx: %w", err)
		return
	}
	PInfo = resp.PInfo
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RQueryServiceConfigEx failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}

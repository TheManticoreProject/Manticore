package functions

// IDL source: [MS-RRASM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrasm/8e6e89fb-9c80-4a9a-a222-d7d8948244bb
// A fetched copy is kept at ms-rrasm.idl in the interface directory.

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rMprAdminServerSetInfoExRequest carries the [in] parameters of RMprAdminServerSetInfoEx.
type rMprAdminServerSetInfoExRequest struct {
	PServerConfig msrrasm.PMPR_SERVER_SET_CONFIG_EX_IDL
}

func (*rMprAdminServerSetInfoExRequest) Opnum() uint16 { return dimsvc.OpnumRMprAdminServerSetInfoEx }

// rMprAdminServerSetInfoExResponse carries the [out] parameters and return value of RMprAdminServerSetInfoEx.
type rMprAdminServerSetInfoExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RMprAdminServerSetInfoEx calls RMprAdminServerSetInfoEx (opnum 47) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RMprAdminServerSetInfoEx(rpc ndr.Invoker, pServerConfig msrrasm.PMPR_SERVER_SET_CONFIG_EX_IDL) (err error) {
	req := &rMprAdminServerSetInfoExRequest{
		PServerConfig: pServerConfig,
	}
	var resp rMprAdminServerSetInfoExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RMprAdminServerSetInfoEx: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RMprAdminServerSetInfoEx failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}

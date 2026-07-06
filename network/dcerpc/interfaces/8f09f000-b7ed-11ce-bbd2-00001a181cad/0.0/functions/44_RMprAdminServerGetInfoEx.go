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

// rMprAdminServerGetInfoExRequest carries the [in] parameters of RMprAdminServerGetInfoEx.
type rMprAdminServerGetInfoExRequest struct {
	PServerConfig msrrasm.PMPR_SERVER_EX_IDL
}

func (*rMprAdminServerGetInfoExRequest) Opnum() uint16 { return dimsvc.OpnumRMprAdminServerGetInfoEx }

// rMprAdminServerGetInfoExResponse carries the [out] parameters and return value of RMprAdminServerGetInfoEx.
type rMprAdminServerGetInfoExResponse struct {
	PServerConfig msrrasm.PMPR_SERVER_EX_IDL
	Status        ndr.DWORD `ndr:"retval"`
}

// RMprAdminServerGetInfoEx calls RMprAdminServerGetInfoEx (opnum 44) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RMprAdminServerGetInfoEx(rpc ndr.Invoker, pServerConfig msrrasm.PMPR_SERVER_EX_IDL) (PServerConfig msrrasm.PMPR_SERVER_EX_IDL, err error) {
	req := &rMprAdminServerGetInfoExRequest{
		PServerConfig: pServerConfig,
	}
	var resp rMprAdminServerGetInfoExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RMprAdminServerGetInfoEx: %w", err)
		return
	}
	PServerConfig = resp.PServerConfig
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RMprAdminServerGetInfoEx failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}

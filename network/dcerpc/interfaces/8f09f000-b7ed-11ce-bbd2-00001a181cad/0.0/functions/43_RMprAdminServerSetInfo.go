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

// rMprAdminServerSetInfoRequest carries the [in] parameters of RMprAdminServerSetInfo.
type rMprAdminServerSetInfoRequest struct {
	DwLevel     ndr.DWORD
	PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER
}

func (*rMprAdminServerSetInfoRequest) Opnum() uint16 { return dimsvc.OpnumRMprAdminServerSetInfo }

// rMprAdminServerSetInfoResponse carries the [out] parameters and return value of RMprAdminServerSetInfo.
type rMprAdminServerSetInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RMprAdminServerSetInfo calls RMprAdminServerSetInfo (opnum 43) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RMprAdminServerSetInfo(rpc ndr.Invoker, dwLevel ndr.DWORD, pInfoStruct msrrasm.DIM_INFORMATION_CONTAINER) (err error) {
	req := &rMprAdminServerSetInfoRequest{
		DwLevel:     dwLevel,
		PInfoStruct: pInfoStruct,
	}
	var resp rMprAdminServerSetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RMprAdminServerSetInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RMprAdminServerSetInfo failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}

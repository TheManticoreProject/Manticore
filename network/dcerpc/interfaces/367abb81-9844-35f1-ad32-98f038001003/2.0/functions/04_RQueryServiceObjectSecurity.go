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

// rQueryServiceObjectSecurityRequest carries the [in] parameters of RQueryServiceObjectSecurity.
type rQueryServiceObjectSecurityRequest struct {
	HService              msscmr.SC_RPC_HANDLE
	DwSecurityInformation ndr.DWORD
	CbBufSize             ndr.DWORD
}

func (*rQueryServiceObjectSecurityRequest) Opnum() uint16 {
	return svcctl.OpnumRQueryServiceObjectSecurity
}

// rQueryServiceObjectSecurityResponse carries the [out] parameters and return value of RQueryServiceObjectSecurity.
type rQueryServiceObjectSecurityResponse struct {
	LpSecurityDescriptor []uint8 `ndr:"ref,size_is=CbBufSize"`
	PcbBytesNeeded       msscmr.LPBOUNDED_DWORD_256K
	Status               ndr.DWORD `ndr:"retval"`
}

// RQueryServiceObjectSecurity calls RQueryServiceObjectSecurity (opnum 4) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RQueryServiceObjectSecurity(rpc ndr.Invoker, hService msscmr.SC_RPC_HANDLE, dwSecurityInformation ndr.DWORD, cbBufSize ndr.DWORD) (LpSecurityDescriptor []uint8, PcbBytesNeeded msscmr.LPBOUNDED_DWORD_256K, err error) {
	req := &rQueryServiceObjectSecurityRequest{
		HService:              hService,
		DwSecurityInformation: dwSecurityInformation,
		CbBufSize:             cbBufSize,
	}
	var resp rQueryServiceObjectSecurityResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RQueryServiceObjectSecurity: %w", err)
		return
	}
	LpSecurityDescriptor = resp.LpSecurityDescriptor
	PcbBytesNeeded = resp.PcbBytesNeeded
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RQueryServiceObjectSecurity failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}

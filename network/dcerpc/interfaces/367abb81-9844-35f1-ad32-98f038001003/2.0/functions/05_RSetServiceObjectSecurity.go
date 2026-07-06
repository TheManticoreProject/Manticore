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

// rSetServiceObjectSecurityRequest carries the [in] parameters of RSetServiceObjectSecurity.
type rSetServiceObjectSecurityRequest struct {
	HService              msscmr.SC_RPC_HANDLE
	DwSecurityInformation ndr.DWORD
	LpSecurityDescriptor  []uint8 `ndr:"ref,size_is=CbBufSize"`
	CbBufSize             ndr.DWORD
}

func (*rSetServiceObjectSecurityRequest) Opnum() uint16 { return svcctl.OpnumRSetServiceObjectSecurity }

// rSetServiceObjectSecurityResponse carries the [out] parameters and return value of RSetServiceObjectSecurity.
type rSetServiceObjectSecurityResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RSetServiceObjectSecurity calls RSetServiceObjectSecurity (opnum 5) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RSetServiceObjectSecurity(rpc ndr.Invoker, hService msscmr.SC_RPC_HANDLE, dwSecurityInformation ndr.DWORD, lpSecurityDescriptor []uint8, cbBufSize ndr.DWORD) (err error) {
	req := &rSetServiceObjectSecurityRequest{
		HService:              hService,
		DwSecurityInformation: dwSecurityInformation,
		LpSecurityDescriptor:  lpSecurityDescriptor,
		CbBufSize:             cbBufSize,
	}
	var resp rSetServiceObjectSecurityResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RSetServiceObjectSecurity: %w", err)
		return
	}
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RSetServiceObjectSecurity failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}

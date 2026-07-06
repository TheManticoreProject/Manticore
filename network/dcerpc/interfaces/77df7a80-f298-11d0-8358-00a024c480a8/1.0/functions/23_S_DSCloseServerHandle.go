package functions

// IDL source: [MS-MQDS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqds/7907bc25-e4e6-40ef-b990-9172d1808e94
// A fetched copy is kept at ms-mqds.idl in the interface directory.

import (
	"fmt"

	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqds "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqds"
)

// s_DSCloseServerHandleRequest carries the [in] parameters of S_DSCloseServerHandle.
type s_DSCloseServerHandleRequest struct {
	PphServerAuth msmqds.PPCONTEXT_HANDLE_SERVER_AUTH_TYPE
}

func (*s_DSCloseServerHandleRequest) Opnum() uint16 { return dscomm.OpnumS_DSCloseServerHandle }

// s_DSCloseServerHandleResponse carries the [out] parameters and return value of S_DSCloseServerHandle.
type s_DSCloseServerHandleResponse struct {
	PphServerAuth msmqds.PPCONTEXT_HANDLE_SERVER_AUTH_TYPE
	Status        ndr.DWORD `ndr:"retval"`
}

// S_DSCloseServerHandle calls S_DSCloseServerHandle (opnum 23) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSCloseServerHandle(rpc ndr.Invoker, pphServerAuth msmqds.PPCONTEXT_HANDLE_SERVER_AUTH_TYPE) (PphServerAuth msmqds.PPCONTEXT_HANDLE_SERVER_AUTH_TYPE, err error) {
	req := &s_DSCloseServerHandleRequest{
		PphServerAuth: pphServerAuth,
	}
	var resp s_DSCloseServerHandleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSCloseServerHandle: %w", err)
		return
	}
	PphServerAuth = resp.PphServerAuth
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSCloseServerHandle failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}

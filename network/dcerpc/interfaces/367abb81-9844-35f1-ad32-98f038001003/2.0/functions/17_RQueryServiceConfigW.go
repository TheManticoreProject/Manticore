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

// rQueryServiceConfigWRequest carries the [in] parameters of RQueryServiceConfigW.
type rQueryServiceConfigWRequest struct {
	HService  msscmr.SC_RPC_HANDLE
	CbBufSize ndr.DWORD
}

func (*rQueryServiceConfigWRequest) Opnum() uint16 { return svcctl.OpnumRQueryServiceConfigW }

// rQueryServiceConfigWResponse carries the [out] parameters and return value of RQueryServiceConfigW.
type rQueryServiceConfigWResponse struct {
	LpServiceConfig msscmr.QUERY_SERVICE_CONFIGW
	PcbBytesNeeded  msscmr.LPBOUNDED_DWORD_8K
	Status          ndr.DWORD `ndr:"retval"`
}

// RQueryServiceConfigW calls RQueryServiceConfigW (opnum 17) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RQueryServiceConfigW(rpc ndr.Invoker, hService msscmr.SC_RPC_HANDLE, cbBufSize ndr.DWORD) (LpServiceConfig msscmr.QUERY_SERVICE_CONFIGW, PcbBytesNeeded msscmr.LPBOUNDED_DWORD_8K, err error) {
	req := &rQueryServiceConfigWRequest{
		HService:  hService,
		CbBufSize: cbBufSize,
	}
	var resp rQueryServiceConfigWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RQueryServiceConfigW: %w", err)
		return
	}
	LpServiceConfig = resp.LpServiceConfig
	PcbBytesNeeded = resp.PcbBytesNeeded
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RQueryServiceConfigW failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}

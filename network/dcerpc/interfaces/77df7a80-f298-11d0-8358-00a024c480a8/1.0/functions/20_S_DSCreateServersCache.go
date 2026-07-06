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

// s_DSCreateServersCacheRequest carries the [in] parameters of S_DSCreateServersCache.
type s_DSCreateServersCacheRequest struct {
	PdwIndex               ndr.DWORD
	LplpSiteServers        *ndr.WSTR `ndr:"unique"`
	PhServerAuth           msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
}

func (*s_DSCreateServersCacheRequest) Opnum() uint16 { return dscomm.OpnumS_DSCreateServersCache }

// s_DSCreateServersCacheResponse carries the [out] parameters and return value of S_DSCreateServersCache.
type s_DSCreateServersCacheResponse struct {
	PdwIndex               ndr.DWORD
	LplpSiteServers        *ndr.WSTR `ndr:"unique"`
	PbServerSignature      []uint8   `ndr:"ref,conformant"`
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
	Status                 ndr.DWORD `ndr:"retval"`
}

// S_DSCreateServersCache calls S_DSCreateServersCache (opnum 20) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSCreateServersCache(rpc ndr.Invoker, pdwIndex ndr.DWORD, lplpSiteServers *ndr.WSTR, phServerAuth msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE, pdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE) (PdwIndex ndr.DWORD, LplpSiteServers *ndr.WSTR, PbServerSignature []uint8, PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE, err error) {
	req := &s_DSCreateServersCacheRequest{
		PdwIndex:               pdwIndex,
		LplpSiteServers:        lplpSiteServers,
		PhServerAuth:           phServerAuth,
		PdwServerSignatureSize: pdwServerSignatureSize,
	}
	var resp s_DSCreateServersCacheResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSCreateServersCache: %w", err)
		return
	}
	PdwIndex = resp.PdwIndex
	LplpSiteServers = resp.LplpSiteServers
	PbServerSignature = resp.PbServerSignature
	PdwServerSignatureSize = resp.PdwServerSignatureSize
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSCreateServersCache failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}

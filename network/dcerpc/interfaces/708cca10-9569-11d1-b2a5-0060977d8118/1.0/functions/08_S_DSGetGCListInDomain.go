package functions

// IDL source: [MS-MQDS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqds/7907bc25-e4e6-40ef-b990-9172d1808e94
// A fetched copy is kept at ms-mqds.idl in the interface directory.

import (
	"fmt"

	dscomm2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/708cca10-9569-11d1-b2a5-0060977d8118/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqds "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqds"
)

// s_DSGetGCListInDomainRequest carries the [in] parameters of S_DSGetGCListInDomain.
type s_DSGetGCListInDomainRequest struct {
	LpwszComputerName      *ndr.WSTR `ndr:"unique"`
	LpwszDomainName        *ndr.WSTR `ndr:"unique"`
	PhServerAuth           msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
}

func (*s_DSGetGCListInDomainRequest) Opnum() uint16 { return dscomm2.OpnumS_DSGetGCListInDomain }

// s_DSGetGCListInDomainResponse carries the [out] parameters and return value of S_DSGetGCListInDomain.
type s_DSGetGCListInDomainResponse struct {
	LplpwszGCList          *ndr.WSTR `ndr:"unique"`
	PbServerSignature      []uint8   `ndr:"ref,conformant"`
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
	Status                 ndr.DWORD `ndr:"retval"`
}

// S_DSGetGCListInDomain calls S_DSGetGCListInDomain (opnum 8) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSGetGCListInDomain(rpc ndr.Invoker, lpwszComputerName *ndr.WSTR, lpwszDomainName *ndr.WSTR, phServerAuth msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE, pdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE) (LplpwszGCList *ndr.WSTR, PbServerSignature []uint8, PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE, err error) {
	req := &s_DSGetGCListInDomainRequest{
		LpwszComputerName:      lpwszComputerName,
		LpwszDomainName:        lpwszDomainName,
		PhServerAuth:           phServerAuth,
		PdwServerSignatureSize: pdwServerSignatureSize,
	}
	var resp s_DSGetGCListInDomainResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSGetGCListInDomain: %w", err)
		return
	}
	LplpwszGCList = resp.LplpwszGCList
	PbServerSignature = resp.PbServerSignature
	PdwServerSignatureSize = resp.PdwServerSignatureSize
	if uint32(resp.Status) != dscomm2.StatusSuccess {
		err = fmt.Errorf("S_DSGetGCListInDomain failed: %s", dscomm2.StatusString(uint32(resp.Status)))
	}
	return
}

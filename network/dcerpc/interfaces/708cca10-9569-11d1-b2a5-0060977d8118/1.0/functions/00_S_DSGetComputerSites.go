package functions

// IDL source: [MS-MQDS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqds/7907bc25-e4e6-40ef-b990-9172d1808e94
// A fetched copy is kept at ms-mqds.idl in the interface directory.

import (
	"fmt"

	dscomm2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/708cca10-9569-11d1-b2a5-0060977d8118/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	msmqds "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqds"
)

// s_DSGetComputerSitesRequest carries the [in] parameters of S_DSGetComputerSites.
type s_DSGetComputerSitesRequest struct {
	PwcsPathName           *ndr.WSTR `ndr:"unique"`
	PhServerAuth           msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
}

func (*s_DSGetComputerSitesRequest) Opnum() uint16 { return dscomm2.OpnumS_DSGetComputerSites }

// s_DSGetComputerSitesResponse carries the [out] parameters and return value of S_DSGetComputerSites.
type s_DSGetComputerSitesResponse struct {
	PdwNumberOfSites       ndr.DWORD
	PpguidSites            []msdtyp.GUID `ndr:"unique,size_is=PdwNumberOfSites,varying,length_is=PdwNumberOfSites"`
	PbServerSignature      []uint8       `ndr:"ref,conformant"`
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
	Status                 ndr.DWORD `ndr:"retval"`
}

// S_DSGetComputerSites calls S_DSGetComputerSites (opnum 0) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSGetComputerSites(rpc ndr.Invoker, pwcsPathName *ndr.WSTR, phServerAuth msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE, pdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE) (PdwNumberOfSites ndr.DWORD, PpguidSites []msdtyp.GUID, PbServerSignature []uint8, PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE, err error) {
	req := &s_DSGetComputerSitesRequest{
		PwcsPathName:           pwcsPathName,
		PhServerAuth:           phServerAuth,
		PdwServerSignatureSize: pdwServerSignatureSize,
	}
	var resp s_DSGetComputerSitesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSGetComputerSites: %w", err)
		return
	}
	PdwNumberOfSites = resp.PdwNumberOfSites
	PpguidSites = resp.PpguidSites
	PbServerSignature = resp.PbServerSignature
	PdwServerSignatureSize = resp.PdwServerSignatureSize
	if uint32(resp.Status) != dscomm2.StatusSuccess {
		err = fmt.Errorf("S_DSGetComputerSites failed: %s", dscomm2.StatusString(uint32(resp.Status)))
	}
	return
}

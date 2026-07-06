package functions

// IDL source: [MS-MQDS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqds/7907bc25-e4e6-40ef-b990-9172d1808e94
// A fetched copy is kept at ms-mqds.idl in the interface directory.

import (
	"fmt"

	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	msmqds "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqds"
)

// s_DSGetObjectSecurityGuidRequest carries the [in] parameters of S_DSGetObjectSecurityGuid.
type s_DSGetObjectSecurityGuidRequest struct {
	DwObjectType           ndr.DWORD
	PGuid                  msdtyp.GUID
	SecurityInformation    ndr.DWORD
	NLength                ndr.DWORD
	PhServerAuth           msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
}

func (*s_DSGetObjectSecurityGuidRequest) Opnum() uint16 { return dscomm.OpnumS_DSGetObjectSecurityGuid }

// s_DSGetObjectSecurityGuidResponse carries the [out] parameters and return value of S_DSGetObjectSecurityGuid.
type s_DSGetObjectSecurityGuidResponse struct {
	PSecurityDescriptor    []uint8 `ndr:"ref,size_is=NLength"`
	LpnLengthNeeded        ndr.DWORD
	PbServerSignature      []uint8 `ndr:"ref,conformant"`
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
	Status                 ndr.DWORD `ndr:"retval"`
}

// S_DSGetObjectSecurityGuid calls S_DSGetObjectSecurityGuid (opnum 13) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSGetObjectSecurityGuid(rpc ndr.Invoker, dwObjectType ndr.DWORD, pGuid msdtyp.GUID, securityInformation ndr.DWORD, nLength ndr.DWORD, phServerAuth msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE, pdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE) (PSecurityDescriptor []uint8, LpnLengthNeeded ndr.DWORD, PbServerSignature []uint8, PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE, err error) {
	req := &s_DSGetObjectSecurityGuidRequest{
		DwObjectType:           dwObjectType,
		PGuid:                  pGuid,
		SecurityInformation:    securityInformation,
		NLength:                nLength,
		PhServerAuth:           phServerAuth,
		PdwServerSignatureSize: pdwServerSignatureSize,
	}
	var resp s_DSGetObjectSecurityGuidResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSGetObjectSecurityGuid: %w", err)
		return
	}
	PSecurityDescriptor = resp.PSecurityDescriptor
	LpnLengthNeeded = resp.LpnLengthNeeded
	PbServerSignature = resp.PbServerSignature
	PdwServerSignatureSize = resp.PdwServerSignatureSize
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSGetObjectSecurityGuid failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}

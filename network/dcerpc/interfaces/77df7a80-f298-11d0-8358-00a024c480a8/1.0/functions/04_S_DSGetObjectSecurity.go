package functions

import (
	"fmt"

	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqds "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqds"
)

// s_DSGetObjectSecurityRequest carries the [in] parameters of S_DSGetObjectSecurity.
type s_DSGetObjectSecurityRequest struct {
	DwObjectType           ndr.DWORD
	PwcsPathName           ndr.WSTR
	SecurityInformation    ndr.DWORD
	NLength                ndr.DWORD
	PhServerAuth           msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
}

func (*s_DSGetObjectSecurityRequest) Opnum() uint16 { return dscomm.OpnumS_DSGetObjectSecurity }

// s_DSGetObjectSecurityResponse carries the [out] parameters and return value of S_DSGetObjectSecurity.
type s_DSGetObjectSecurityResponse struct {
	PSecurityDescriptor    []uint8 `ndr:"ref,size_is=NLength"`
	LpnLengthNeeded        ndr.DWORD
	PbServerSignature      []uint8 `ndr:"ref,conformant"`
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
	Status                 ndr.DWORD `ndr:"retval"`
}

// S_DSGetObjectSecurity calls S_DSGetObjectSecurity (opnum 4) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSGetObjectSecurity(rpc ndr.Invoker, dwObjectType ndr.DWORD, pwcsPathName ndr.WSTR, securityInformation ndr.DWORD, nLength ndr.DWORD, phServerAuth msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE, pdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE) (PSecurityDescriptor []uint8, LpnLengthNeeded ndr.DWORD, PbServerSignature []uint8, PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE, err error) {
	req := &s_DSGetObjectSecurityRequest{
		DwObjectType:           dwObjectType,
		PwcsPathName:           pwcsPathName,
		SecurityInformation:    securityInformation,
		NLength:                nLength,
		PhServerAuth:           phServerAuth,
		PdwServerSignatureSize: pdwServerSignatureSize,
	}
	var resp s_DSGetObjectSecurityResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSGetObjectSecurity: %w", err)
		return
	}
	PSecurityDescriptor = resp.PSecurityDescriptor
	LpnLengthNeeded = resp.LpnLengthNeeded
	PbServerSignature = resp.PbServerSignature
	PdwServerSignatureSize = resp.PdwServerSignatureSize
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSGetObjectSecurity failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}

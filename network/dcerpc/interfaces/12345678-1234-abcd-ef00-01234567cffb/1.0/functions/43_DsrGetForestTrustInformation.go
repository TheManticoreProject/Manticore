package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// dsrGetForestTrustInformationRequest carries the [in] parameters of DsrGetForestTrustInformation.
type dsrGetForestTrustInformationRequest struct {
	ServerName        *ndr.WSTR `ndr:"unique"`
	TrustedDomainName *ndr.WSTR `ndr:"unique"`
	Flags             ndr.DWORD
}

func (*dsrGetForestTrustInformationRequest) Opnum() uint16 {
	return logon.OpnumDsrGetForestTrustInformation
}

// dsrGetForestTrustInformationResponse carries the [out] parameters and return value of DsrGetForestTrustInformation.
type dsrGetForestTrustInformationResponse struct {
	ForestTrustInfo *msnrpc.LSA_FOREST_TRUST_INFORMATION `ndr:"unique"`
	Status          ndr.DWORD                            `ndr:"retval"`
}

// DsrGetForestTrustInformation calls DsrGetForestTrustInformation (opnum 43) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func DsrGetForestTrustInformation(rpc ndr.Invoker, serverName *ndr.WSTR, trustedDomainName *ndr.WSTR, flags ndr.DWORD) (ForestTrustInfo *msnrpc.LSA_FOREST_TRUST_INFORMATION, err error) {
	req := &dsrGetForestTrustInformationRequest{
		ServerName:        serverName,
		TrustedDomainName: trustedDomainName,
		Flags:             flags,
	}
	var resp dsrGetForestTrustInformationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("DsrGetForestTrustInformation: %w", err)
		return
	}
	ForestTrustInfo = resp.ForestTrustInfo
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("DsrGetForestTrustInformation failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

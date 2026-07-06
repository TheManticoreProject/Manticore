package functions

// IDL source: [MS-NRPC] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7
// A fetched copy is kept at ms-nrpc.idl in the interface directory.

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// netrGetForestTrustInformationRequest carries the [in] parameters of NetrGetForestTrustInformation.
type netrGetForestTrustInformationRequest struct {
	ServerName    *ndr.WSTR `ndr:"unique"`
	ComputerName  ndr.WSTR
	Authenticator msnrpc.NETLOGON_AUTHENTICATOR
	Flags         ndr.DWORD
}

func (*netrGetForestTrustInformationRequest) Opnum() uint16 {
	return logon.OpnumNetrGetForestTrustInformation
}

// netrGetForestTrustInformationResponse carries the [out] parameters and return value of NetrGetForestTrustInformation.
type netrGetForestTrustInformationResponse struct {
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	ForestTrustInfo     *msnrpc.LSA_FOREST_TRUST_INFORMATION `ndr:"unique"`
	Status              ndr.DWORD                            `ndr:"retval"`
}

// NetrGetForestTrustInformation calls NetrGetForestTrustInformation (opnum 44) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrGetForestTrustInformation(rpc ndr.Invoker, serverName *ndr.WSTR, computerName ndr.WSTR, authenticator msnrpc.NETLOGON_AUTHENTICATOR, flags ndr.DWORD) (ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, ForestTrustInfo *msnrpc.LSA_FOREST_TRUST_INFORMATION, err error) {
	req := &netrGetForestTrustInformationRequest{
		ServerName:    serverName,
		ComputerName:  computerName,
		Authenticator: authenticator,
		Flags:         flags,
	}
	var resp netrGetForestTrustInformationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrGetForestTrustInformation: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	ForestTrustInfo = resp.ForestTrustInfo
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrGetForestTrustInformation failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

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

// netrServerTrustPasswordsGetRequest carries the [in] parameters of NetrServerTrustPasswordsGet.
type netrServerTrustPasswordsGetRequest struct {
	TrustedDcName     *ndr.WSTR `ndr:"unique"`
	AccountName       ndr.WSTR
	SecureChannelType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE
	ComputerName      ndr.WSTR
	Authenticator     msnrpc.NETLOGON_AUTHENTICATOR
}

func (*netrServerTrustPasswordsGetRequest) Opnum() uint16 {
	return logon.OpnumNetrServerTrustPasswordsGet
}

// netrServerTrustPasswordsGetResponse carries the [out] parameters and return value of NetrServerTrustPasswordsGet.
type netrServerTrustPasswordsGetResponse struct {
	ReturnAuthenticator     msnrpc.NETLOGON_AUTHENTICATOR
	EncryptedNewOwfPassword msnrpc.NT_OWF_PASSWORD
	EncryptedOldOwfPassword msnrpc.NT_OWF_PASSWORD
	Status                  ndr.DWORD `ndr:"retval"`
}

// NetrServerTrustPasswordsGet calls NetrServerTrustPasswordsGet (opnum 42) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrServerTrustPasswordsGet(rpc ndr.Invoker, trustedDcName *ndr.WSTR, accountName ndr.WSTR, secureChannelType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE, computerName ndr.WSTR, authenticator msnrpc.NETLOGON_AUTHENTICATOR) (ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, EncryptedNewOwfPassword msnrpc.NT_OWF_PASSWORD, EncryptedOldOwfPassword msnrpc.NT_OWF_PASSWORD, err error) {
	req := &netrServerTrustPasswordsGetRequest{
		TrustedDcName:     trustedDcName,
		AccountName:       accountName,
		SecureChannelType: secureChannelType,
		ComputerName:      computerName,
		Authenticator:     authenticator,
	}
	var resp netrServerTrustPasswordsGetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrServerTrustPasswordsGet: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	EncryptedNewOwfPassword = resp.EncryptedNewOwfPassword
	EncryptedOldOwfPassword = resp.EncryptedOldOwfPassword
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrServerTrustPasswordsGet failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

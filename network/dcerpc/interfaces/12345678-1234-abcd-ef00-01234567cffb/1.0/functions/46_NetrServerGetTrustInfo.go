package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// netrServerGetTrustInfoRequest carries the [in] parameters of NetrServerGetTrustInfo.
type netrServerGetTrustInfoRequest struct {
	TrustedDcName     *ndr.WSTR `ndr:"unique"`
	AccountName       ndr.WSTR
	SecureChannelType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE
	ComputerName      ndr.WSTR
	Authenticator     msnrpc.NETLOGON_AUTHENTICATOR
}

func (*netrServerGetTrustInfoRequest) Opnum() uint16 { return logon.OpnumNetrServerGetTrustInfo }

// netrServerGetTrustInfoResponse carries the [out] parameters and return value of NetrServerGetTrustInfo.
type netrServerGetTrustInfoResponse struct {
	ReturnAuthenticator     msnrpc.NETLOGON_AUTHENTICATOR
	EncryptedNewOwfPassword msnrpc.NT_OWF_PASSWORD
	EncryptedOldOwfPassword msnrpc.NT_OWF_PASSWORD
	TrustInfo               *msnrpc.NL_GENERIC_RPC_DATA `ndr:"unique"`
	Status                  ndr.DWORD                   `ndr:"retval"`
}

// NetrServerGetTrustInfo calls NetrServerGetTrustInfo (opnum 46) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrServerGetTrustInfo(rpc ndr.Invoker, trustedDcName *ndr.WSTR, accountName ndr.WSTR, secureChannelType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE, computerName ndr.WSTR, authenticator msnrpc.NETLOGON_AUTHENTICATOR) (ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, EncryptedNewOwfPassword msnrpc.NT_OWF_PASSWORD, EncryptedOldOwfPassword msnrpc.NT_OWF_PASSWORD, TrustInfo *msnrpc.NL_GENERIC_RPC_DATA, err error) {
	req := &netrServerGetTrustInfoRequest{
		TrustedDcName:     trustedDcName,
		AccountName:       accountName,
		SecureChannelType: secureChannelType,
		ComputerName:      computerName,
		Authenticator:     authenticator,
	}
	var resp netrServerGetTrustInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrServerGetTrustInfo: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	EncryptedNewOwfPassword = resp.EncryptedNewOwfPassword
	EncryptedOldOwfPassword = resp.EncryptedOldOwfPassword
	TrustInfo = resp.TrustInfo
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrServerGetTrustInfo failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

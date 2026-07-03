package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// netrServerAuthenticateRequest carries the [in] parameters of NetrServerAuthenticate.
type netrServerAuthenticateRequest struct {
	PrimaryName       *ndr.WSTR `ndr:"unique"`
	AccountName       ndr.WSTR
	SecureChannelType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE
	ComputerName      ndr.WSTR
	ClientCredential  msnrpc.NETLOGON_CREDENTIAL
}

func (*netrServerAuthenticateRequest) Opnum() uint16 { return logon.OpnumNetrServerAuthenticate }

// netrServerAuthenticateResponse carries the [out] parameters and return value of NetrServerAuthenticate.
type netrServerAuthenticateResponse struct {
	ServerCredential msnrpc.NETLOGON_CREDENTIAL
	Status           ndr.DWORD `ndr:"retval"`
}

// NetrServerAuthenticate calls NetrServerAuthenticate (opnum 5) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrServerAuthenticate(rpc ndr.Invoker, primaryName *ndr.WSTR, accountName ndr.WSTR, secureChannelType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE, computerName ndr.WSTR, clientCredential msnrpc.NETLOGON_CREDENTIAL) (ServerCredential msnrpc.NETLOGON_CREDENTIAL, err error) {
	req := &netrServerAuthenticateRequest{
		PrimaryName:       primaryName,
		AccountName:       accountName,
		SecureChannelType: secureChannelType,
		ComputerName:      computerName,
		ClientCredential:  clientCredential,
	}
	var resp netrServerAuthenticateResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrServerAuthenticate: %w", err)
		return
	}
	ServerCredential = resp.ServerCredential
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrServerAuthenticate failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

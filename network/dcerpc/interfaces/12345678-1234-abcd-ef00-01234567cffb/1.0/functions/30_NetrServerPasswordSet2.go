package functions

import (
	"fmt"

	netlogon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

type netrServerPasswordSet2Request struct {
	PrimaryName       *ndr.WSTR `ndr:"unique"`
	AccountName       ndr.WSTR
	SecureChannelType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE `ndr:"enum"`
	ComputerName      ndr.WSTR
	Authenticator     msnrpc.NETLOGON_AUTHENTICATOR
	ClearNewPassword  msnrpc.NL_TRUST_PASSWORD
}

func (*netrServerPasswordSet2Request) Opnum() uint16 {
	return netlogon.OpnumNetrServerPasswordSet2
}

type netrServerPasswordSet2Response struct {
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	Status              ndr.DWORD `ndr:"retval"`
}

// NetrServerPasswordSet2 calls NetrServerPasswordSet2 ([MS-NRPC] 3.5.4.4.5, opnum 30) to
// set the account password. Passing an all-zero ClearNewPassword blanks the target
// account's password.
//
// Parameters:
//   - rpc: A DCE/RPC client bound to the Netlogon interface.
//   - primaryName: The DC name (empty sends a NULL unique pointer).
//   - accountName: The account whose password is being set (e.g. "DC$").
//   - secureChannelType: The secure channel kind.
//   - computerName: The NetBIOS name of the client computer.
//   - authenticator: The request authenticator.
//   - clearNewPassword: The new password structure.
//
// Returns:
//   - The return authenticator, the NTSTATUS, and a transport error.
func NetrServerPasswordSet2(rpc ndr.Invoker, primaryName, accountName string, secureChannelType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE, computerName string, authenticator msnrpc.NETLOGON_AUTHENTICATOR, clearNewPassword msnrpc.NL_TRUST_PASSWORD) (msnrpc.NETLOGON_AUTHENTICATOR, uint32, error) {
	req := &netrServerPasswordSet2Request{
		AccountName:       ndr.WSTR(accountName),
		SecureChannelType: secureChannelType,
		ComputerName:      ndr.WSTR(computerName),
		Authenticator:     authenticator,
		ClearNewPassword:  clearNewPassword,
	}
	if primaryName != "" {
		pn := ndr.WSTR(primaryName)
		req.PrimaryName = &pn
	}

	var resp netrServerPasswordSet2Response
	if err := rpc.Invoke(req, &resp); err != nil {
		return msnrpc.NETLOGON_AUTHENTICATOR{}, 0, fmt.Errorf("NetrServerPasswordSet2: %w", err)
	}
	return resp.ReturnAuthenticator, uint32(resp.Status), nil
}

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

// netrServerPasswordSetRequest carries the [in] parameters of NetrServerPasswordSet.
type netrServerPasswordSetRequest struct {
	PrimaryName       *ndr.WSTR `ndr:"unique"`
	AccountName       ndr.WSTR
	SecureChannelType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE
	ComputerName      ndr.WSTR
	Authenticator     msnrpc.NETLOGON_AUTHENTICATOR
	UasNewPassword    msnrpc.NT_OWF_PASSWORD
}

func (*netrServerPasswordSetRequest) Opnum() uint16 { return logon.OpnumNetrServerPasswordSet }

// netrServerPasswordSetResponse carries the [out] parameters and return value of NetrServerPasswordSet.
type netrServerPasswordSetResponse struct {
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	Status              ndr.DWORD `ndr:"retval"`
}

// NetrServerPasswordSet calls NetrServerPasswordSet (opnum 6) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrServerPasswordSet(rpc ndr.Invoker, primaryName *ndr.WSTR, accountName ndr.WSTR, secureChannelType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE, computerName ndr.WSTR, authenticator msnrpc.NETLOGON_AUTHENTICATOR, uasNewPassword msnrpc.NT_OWF_PASSWORD) (ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, err error) {
	req := &netrServerPasswordSetRequest{
		PrimaryName:       primaryName,
		AccountName:       accountName,
		SecureChannelType: secureChannelType,
		ComputerName:      computerName,
		Authenticator:     authenticator,
		UasNewPassword:    uasNewPassword,
	}
	var resp netrServerPasswordSetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrServerPasswordSet: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrServerPasswordSet failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

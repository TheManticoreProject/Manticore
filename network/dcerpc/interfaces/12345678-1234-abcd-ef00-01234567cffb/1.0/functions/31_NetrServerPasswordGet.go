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

// netrServerPasswordGetRequest carries the [in] parameters of NetrServerPasswordGet.
type netrServerPasswordGetRequest struct {
	PrimaryName   *ndr.WSTR `ndr:"unique"`
	AccountName   ndr.WSTR
	AccountType   msnrpc.NETLOGON_SECURE_CHANNEL_TYPE
	ComputerName  ndr.WSTR
	Authenticator msnrpc.NETLOGON_AUTHENTICATOR
}

func (*netrServerPasswordGetRequest) Opnum() uint16 { return logon.OpnumNetrServerPasswordGet }

// netrServerPasswordGetResponse carries the [out] parameters and return value of NetrServerPasswordGet.
type netrServerPasswordGetResponse struct {
	ReturnAuthenticator    msnrpc.NETLOGON_AUTHENTICATOR
	EncryptedNtOwfPassword msnrpc.NT_OWF_PASSWORD
	Status                 ndr.DWORD `ndr:"retval"`
}

// NetrServerPasswordGet calls NetrServerPasswordGet (opnum 31) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrServerPasswordGet(rpc ndr.Invoker, primaryName *ndr.WSTR, accountName ndr.WSTR, accountType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE, computerName ndr.WSTR, authenticator msnrpc.NETLOGON_AUTHENTICATOR) (ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, EncryptedNtOwfPassword msnrpc.NT_OWF_PASSWORD, err error) {
	req := &netrServerPasswordGetRequest{
		PrimaryName:   primaryName,
		AccountName:   accountName,
		AccountType:   accountType,
		ComputerName:  computerName,
		Authenticator: authenticator,
	}
	var resp netrServerPasswordGetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrServerPasswordGet: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	EncryptedNtOwfPassword = resp.EncryptedNtOwfPassword
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrServerPasswordGet failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

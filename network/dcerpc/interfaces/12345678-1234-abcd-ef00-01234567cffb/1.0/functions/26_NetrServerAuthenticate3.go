package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// netrServerAuthenticate3Request carries the [in] parameters of NetrServerAuthenticate3.
type netrServerAuthenticate3Request struct {
	PrimaryName       *ndr.WSTR `ndr:"unique"`
	AccountName       ndr.WSTR
	SecureChannelType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE
	ComputerName      ndr.WSTR
	ClientCredential  msnrpc.NETLOGON_CREDENTIAL
	NegotiateFlags    ndr.DWORD
}

func (*netrServerAuthenticate3Request) Opnum() uint16 { return logon.OpnumNetrServerAuthenticate3 }

// netrServerAuthenticate3Response carries the [out] parameters and return value of NetrServerAuthenticate3.
type netrServerAuthenticate3Response struct {
	ServerCredential msnrpc.NETLOGON_CREDENTIAL
	NegotiateFlags   ndr.DWORD
	AccountRid       ndr.DWORD
	Status           ndr.DWORD `ndr:"retval"`
}

// NetrServerAuthenticate3 calls NetrServerAuthenticate3 (opnum 26) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrServerAuthenticate3(rpc ndr.Invoker, primaryName *ndr.WSTR, accountName ndr.WSTR, secureChannelType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE, computerName ndr.WSTR, clientCredential msnrpc.NETLOGON_CREDENTIAL, negotiateFlags ndr.DWORD) (ServerCredential msnrpc.NETLOGON_CREDENTIAL, NegotiateFlags ndr.DWORD, AccountRid ndr.DWORD, err error) {
	req := &netrServerAuthenticate3Request{
		PrimaryName:       primaryName,
		AccountName:       accountName,
		SecureChannelType: secureChannelType,
		ComputerName:      computerName,
		ClientCredential:  clientCredential,
		NegotiateFlags:    negotiateFlags,
	}
	var resp netrServerAuthenticate3Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrServerAuthenticate3: %w", err)
		return
	}
	ServerCredential = resp.ServerCredential
	NegotiateFlags = resp.NegotiateFlags
	AccountRid = resp.AccountRid
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrServerAuthenticate3 failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// netrServerAuthenticateKerberosRequest carries the [in] parameters of NetrServerAuthenticateKerberos.
type netrServerAuthenticateKerberosRequest struct {
	PrimaryName    *ndr.WSTR `ndr:"unique"`
	AccountName    ndr.WSTR
	AccountType    msnrpc.NETLOGON_SECURE_CHANNEL_TYPE
	ComputerName   ndr.WSTR
	NegotiateFlags ndr.DWORD
}

func (*netrServerAuthenticateKerberosRequest) Opnum() uint16 {
	return logon.OpnumNetrServerAuthenticateKerberos
}

// netrServerAuthenticateKerberosResponse carries the [out] parameters and return value of NetrServerAuthenticateKerberos.
type netrServerAuthenticateKerberosResponse struct {
	NegotiateFlags ndr.DWORD
	AccountRid     ndr.DWORD
	Status         ndr.DWORD `ndr:"retval"`
}

// NetrServerAuthenticateKerberos calls NetrServerAuthenticateKerberos (opnum 59) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrServerAuthenticateKerberos(rpc ndr.Invoker, primaryName *ndr.WSTR, accountName ndr.WSTR, accountType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE, computerName ndr.WSTR, negotiateFlags ndr.DWORD) (NegotiateFlags ndr.DWORD, AccountRid ndr.DWORD, err error) {
	req := &netrServerAuthenticateKerberosRequest{
		PrimaryName:    primaryName,
		AccountName:    accountName,
		AccountType:    accountType,
		ComputerName:   computerName,
		NegotiateFlags: negotiateFlags,
	}
	var resp netrServerAuthenticateKerberosResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrServerAuthenticateKerberos: %w", err)
		return
	}
	NegotiateFlags = resp.NegotiateFlags
	AccountRid = resp.AccountRid
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrServerAuthenticateKerberos failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

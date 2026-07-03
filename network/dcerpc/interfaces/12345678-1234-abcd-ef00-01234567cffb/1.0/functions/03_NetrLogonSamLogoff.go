package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// netrLogonSamLogoffRequest carries the [in] parameters of NetrLogonSamLogoff.
type netrLogonSamLogoffRequest struct {
	LogonServer         *ndr.WSTR                      `ndr:"unique"`
	ComputerName        *ndr.WSTR                      `ndr:"unique"`
	Authenticator       *msnrpc.NETLOGON_AUTHENTICATOR `ndr:"unique"`
	ReturnAuthenticator *msnrpc.NETLOGON_AUTHENTICATOR `ndr:"unique"`
	LogonLevel          msnrpc.NETLOGON_LOGON_INFO_CLASS
	LogonInformation    msnrpc.NETLOGON_LEVEL
}

func (*netrLogonSamLogoffRequest) Opnum() uint16 { return logon.OpnumNetrLogonSamLogoff }

// netrLogonSamLogoffResponse carries the [out] parameters and return value of NetrLogonSamLogoff.
type netrLogonSamLogoffResponse struct {
	ReturnAuthenticator *msnrpc.NETLOGON_AUTHENTICATOR `ndr:"unique"`
	Status              ndr.DWORD                      `ndr:"retval"`
}

// NetrLogonSamLogoff calls NetrLogonSamLogoff (opnum 3) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrLogonSamLogoff(rpc ndr.Invoker, logonServer *ndr.WSTR, computerName *ndr.WSTR, authenticator *msnrpc.NETLOGON_AUTHENTICATOR, returnAuthenticator *msnrpc.NETLOGON_AUTHENTICATOR, logonLevel msnrpc.NETLOGON_LOGON_INFO_CLASS, logonInformation msnrpc.NETLOGON_LEVEL) (ReturnAuthenticator *msnrpc.NETLOGON_AUTHENTICATOR, err error) {
	req := &netrLogonSamLogoffRequest{
		LogonServer:         logonServer,
		ComputerName:        computerName,
		Authenticator:       authenticator,
		ReturnAuthenticator: returnAuthenticator,
		LogonLevel:          logonLevel,
		LogonInformation:    logonInformation,
	}
	var resp netrLogonSamLogoffResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrLogonSamLogoff: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrLogonSamLogoff failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

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

// netrLogonGetCapabilitiesRequest carries the [in] parameters of NetrLogonGetCapabilities.
type netrLogonGetCapabilitiesRequest struct {
	ServerName          ndr.WSTR
	ComputerName        *ndr.WSTR `ndr:"unique"`
	Authenticator       msnrpc.NETLOGON_AUTHENTICATOR
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	QueryLevel          ndr.DWORD
}

func (*netrLogonGetCapabilitiesRequest) Opnum() uint16 { return logon.OpnumNetrLogonGetCapabilities }

// netrLogonGetCapabilitiesResponse carries the [out] parameters and return value of NetrLogonGetCapabilities.
type netrLogonGetCapabilitiesResponse struct {
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	Capabilities        msnrpc.NETLOGON_CAPABILITIES
	Status              ndr.DWORD `ndr:"retval"`
}

// NetrLogonGetCapabilities calls NetrLogonGetCapabilities (opnum 21) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrLogonGetCapabilities(rpc ndr.Invoker, serverName ndr.WSTR, computerName *ndr.WSTR, authenticator msnrpc.NETLOGON_AUTHENTICATOR, returnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, queryLevel ndr.DWORD) (ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, Capabilities msnrpc.NETLOGON_CAPABILITIES, err error) {
	req := &netrLogonGetCapabilitiesRequest{
		ServerName:          serverName,
		ComputerName:        computerName,
		Authenticator:       authenticator,
		ReturnAuthenticator: returnAuthenticator,
		QueryLevel:          queryLevel,
	}
	var resp netrLogonGetCapabilitiesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrLogonGetCapabilities: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	Capabilities = resp.Capabilities
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrLogonGetCapabilities failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

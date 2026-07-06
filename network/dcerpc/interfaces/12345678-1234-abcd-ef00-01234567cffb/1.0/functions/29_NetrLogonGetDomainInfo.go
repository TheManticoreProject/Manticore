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

// netrLogonGetDomainInfoRequest carries the [in] parameters of NetrLogonGetDomainInfo.
type netrLogonGetDomainInfoRequest struct {
	ServerName          ndr.WSTR
	ComputerName        *ndr.WSTR `ndr:"unique"`
	Authenticator       msnrpc.NETLOGON_AUTHENTICATOR
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	Level               ndr.DWORD
	WkstaBuffer         msnrpc.NETLOGON_WORKSTATION_INFORMATION
}

func (*netrLogonGetDomainInfoRequest) Opnum() uint16 { return logon.OpnumNetrLogonGetDomainInfo }

// netrLogonGetDomainInfoResponse carries the [out] parameters and return value of NetrLogonGetDomainInfo.
type netrLogonGetDomainInfoResponse struct {
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	DomBuffer           msnrpc.NETLOGON_DOMAIN_INFORMATION
	Status              ndr.DWORD `ndr:"retval"`
}

// NetrLogonGetDomainInfo calls NetrLogonGetDomainInfo (opnum 29) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrLogonGetDomainInfo(rpc ndr.Invoker, serverName ndr.WSTR, computerName *ndr.WSTR, authenticator msnrpc.NETLOGON_AUTHENTICATOR, returnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, level ndr.DWORD, wkstaBuffer msnrpc.NETLOGON_WORKSTATION_INFORMATION) (ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, DomBuffer msnrpc.NETLOGON_DOMAIN_INFORMATION, err error) {
	req := &netrLogonGetDomainInfoRequest{
		ServerName:          serverName,
		ComputerName:        computerName,
		Authenticator:       authenticator,
		ReturnAuthenticator: returnAuthenticator,
		Level:               level,
		WkstaBuffer:         wkstaBuffer,
	}
	var resp netrLogonGetDomainInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrLogonGetDomainInfo: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	DomBuffer = resp.DomBuffer
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrLogonGetDomainInfo failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

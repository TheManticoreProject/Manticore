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

// netrEnumerateTrustedDomainsRequest carries the [in] parameters of NetrEnumerateTrustedDomains.
type netrEnumerateTrustedDomainsRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
}

func (*netrEnumerateTrustedDomainsRequest) Opnum() uint16 {
	return logon.OpnumNetrEnumerateTrustedDomains
}

// netrEnumerateTrustedDomainsResponse carries the [out] parameters and return value of NetrEnumerateTrustedDomains.
type netrEnumerateTrustedDomainsResponse struct {
	DomainNameBuffer msnrpc.DOMAIN_NAME_BUFFER
	Status           ndr.DWORD `ndr:"retval"`
}

// NetrEnumerateTrustedDomains calls NetrEnumerateTrustedDomains (opnum 19) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrEnumerateTrustedDomains(rpc ndr.Invoker, serverName *ndr.WSTR) (DomainNameBuffer msnrpc.DOMAIN_NAME_BUFFER, err error) {
	req := &netrEnumerateTrustedDomainsRequest{
		ServerName: serverName,
	}
	var resp netrEnumerateTrustedDomainsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrEnumerateTrustedDomains: %w", err)
		return
	}
	DomainNameBuffer = resp.DomainNameBuffer
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrEnumerateTrustedDomains failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

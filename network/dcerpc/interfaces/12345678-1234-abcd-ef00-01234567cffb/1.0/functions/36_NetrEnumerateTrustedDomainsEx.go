package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// netrEnumerateTrustedDomainsExRequest carries the [in] parameters of NetrEnumerateTrustedDomainsEx.
type netrEnumerateTrustedDomainsExRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
}

func (*netrEnumerateTrustedDomainsExRequest) Opnum() uint16 {
	return logon.OpnumNetrEnumerateTrustedDomainsEx
}

// netrEnumerateTrustedDomainsExResponse carries the [out] parameters and return value of NetrEnumerateTrustedDomainsEx.
type netrEnumerateTrustedDomainsExResponse struct {
	Domains msnrpc.NETLOGON_TRUSTED_DOMAIN_ARRAY
	Status  ndr.DWORD `ndr:"retval"`
}

// NetrEnumerateTrustedDomainsEx calls NetrEnumerateTrustedDomainsEx (opnum 36) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrEnumerateTrustedDomainsEx(rpc ndr.Invoker, serverName *ndr.WSTR) (Domains msnrpc.NETLOGON_TRUSTED_DOMAIN_ARRAY, err error) {
	req := &netrEnumerateTrustedDomainsExRequest{
		ServerName: serverName,
	}
	var resp netrEnumerateTrustedDomainsExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrEnumerateTrustedDomainsEx: %w", err)
		return
	}
	Domains = resp.Domains
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrEnumerateTrustedDomainsEx failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

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

// dsrEnumerateDomainTrustsRequest carries the [in] parameters of DsrEnumerateDomainTrusts.
type dsrEnumerateDomainTrustsRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Flags      ndr.DWORD
}

func (*dsrEnumerateDomainTrustsRequest) Opnum() uint16 { return logon.OpnumDsrEnumerateDomainTrusts }

// dsrEnumerateDomainTrustsResponse carries the [out] parameters and return value of DsrEnumerateDomainTrusts.
type dsrEnumerateDomainTrustsResponse struct {
	Domains msnrpc.NETLOGON_TRUSTED_DOMAIN_ARRAY
	Status  ndr.DWORD `ndr:"retval"`
}

// DsrEnumerateDomainTrusts calls DsrEnumerateDomainTrusts (opnum 40) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func DsrEnumerateDomainTrusts(rpc ndr.Invoker, serverName *ndr.WSTR, flags ndr.DWORD) (Domains msnrpc.NETLOGON_TRUSTED_DOMAIN_ARRAY, err error) {
	req := &dsrEnumerateDomainTrustsRequest{
		ServerName: serverName,
		Flags:      flags,
	}
	var resp dsrEnumerateDomainTrustsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("DsrEnumerateDomainTrusts: %w", err)
		return
	}
	Domains = resp.Domains
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("DsrEnumerateDomainTrusts failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

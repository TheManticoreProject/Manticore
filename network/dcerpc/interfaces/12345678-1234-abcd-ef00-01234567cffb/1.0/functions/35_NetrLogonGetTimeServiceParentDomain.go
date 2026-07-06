package functions

// IDL source: [MS-NRPC] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7
// A fetched copy is kept at ms-nrpc.idl in the interface directory.

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrLogonGetTimeServiceParentDomainRequest carries the [in] parameters of NetrLogonGetTimeServiceParentDomain.
type netrLogonGetTimeServiceParentDomainRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
}

func (*netrLogonGetTimeServiceParentDomainRequest) Opnum() uint16 {
	return logon.OpnumNetrLogonGetTimeServiceParentDomain
}

// netrLogonGetTimeServiceParentDomainResponse carries the [out] parameters and return value of NetrLogonGetTimeServiceParentDomain.
type netrLogonGetTimeServiceParentDomainResponse struct {
	DomainName  ndr.WSTR
	PdcSameSite int32
	Status      ndr.DWORD `ndr:"retval"`
}

// NetrLogonGetTimeServiceParentDomain calls NetrLogonGetTimeServiceParentDomain (opnum 35) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrLogonGetTimeServiceParentDomain(rpc ndr.Invoker, serverName *ndr.WSTR) (DomainName ndr.WSTR, PdcSameSite int32, err error) {
	req := &netrLogonGetTimeServiceParentDomainRequest{
		ServerName: serverName,
	}
	var resp netrLogonGetTimeServiceParentDomainResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrLogonGetTimeServiceParentDomain: %w", err)
		return
	}
	DomainName = resp.DomainName
	PdcSameSite = resp.PdcSameSite
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrLogonGetTimeServiceParentDomain failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

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

// netrLogonGetTrustRidRequest carries the [in] parameters of NetrLogonGetTrustRid.
type netrLogonGetTrustRidRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	DomainName *ndr.WSTR `ndr:"unique"`
}

func (*netrLogonGetTrustRidRequest) Opnum() uint16 { return logon.OpnumNetrLogonGetTrustRid }

// netrLogonGetTrustRidResponse carries the [out] parameters and return value of NetrLogonGetTrustRid.
type netrLogonGetTrustRidResponse struct {
	Rid    ndr.DWORD
	Status ndr.DWORD `ndr:"retval"`
}

// NetrLogonGetTrustRid calls NetrLogonGetTrustRid (opnum 23) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrLogonGetTrustRid(rpc ndr.Invoker, serverName *ndr.WSTR, domainName *ndr.WSTR) (Rid ndr.DWORD, err error) {
	req := &netrLogonGetTrustRidRequest{
		ServerName: serverName,
		DomainName: domainName,
	}
	var resp netrLogonGetTrustRidResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrLogonGetTrustRid: %w", err)
		return
	}
	Rid = resp.Rid
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrLogonGetTrustRid failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

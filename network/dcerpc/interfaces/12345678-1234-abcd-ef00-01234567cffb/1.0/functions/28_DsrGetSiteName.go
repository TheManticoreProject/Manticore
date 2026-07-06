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

// dsrGetSiteNameRequest carries the [in] parameters of DsrGetSiteName.
type dsrGetSiteNameRequest struct {
	ComputerName *ndr.WSTR `ndr:"unique"`
}

func (*dsrGetSiteNameRequest) Opnum() uint16 { return logon.OpnumDsrGetSiteName }

// dsrGetSiteNameResponse carries the [out] parameters and return value of DsrGetSiteName.
type dsrGetSiteNameResponse struct {
	SiteName ndr.WSTR
	Status   ndr.DWORD `ndr:"retval"`
}

// DsrGetSiteName calls DsrGetSiteName (opnum 28) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func DsrGetSiteName(rpc ndr.Invoker, computerName *ndr.WSTR) (SiteName ndr.WSTR, err error) {
	req := &dsrGetSiteNameRequest{
		ComputerName: computerName,
	}
	var resp dsrGetSiteNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("DsrGetSiteName: %w", err)
		return
	}
	SiteName = resp.SiteName
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("DsrGetSiteName failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

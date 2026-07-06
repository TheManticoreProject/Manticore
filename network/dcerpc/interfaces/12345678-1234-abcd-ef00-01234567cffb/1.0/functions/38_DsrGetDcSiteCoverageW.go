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

// dsrGetDcSiteCoverageWRequest carries the [in] parameters of DsrGetDcSiteCoverageW.
type dsrGetDcSiteCoverageWRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
}

func (*dsrGetDcSiteCoverageWRequest) Opnum() uint16 { return logon.OpnumDsrGetDcSiteCoverageW }

// dsrGetDcSiteCoverageWResponse carries the [out] parameters and return value of DsrGetDcSiteCoverageW.
type dsrGetDcSiteCoverageWResponse struct {
	SiteNames *msnrpc.NL_SITE_NAME_ARRAY `ndr:"unique"`
	Status    ndr.DWORD                  `ndr:"retval"`
}

// DsrGetDcSiteCoverageW calls DsrGetDcSiteCoverageW (opnum 38) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func DsrGetDcSiteCoverageW(rpc ndr.Invoker, serverName *ndr.WSTR) (SiteNames *msnrpc.NL_SITE_NAME_ARRAY, err error) {
	req := &dsrGetDcSiteCoverageWRequest{
		ServerName: serverName,
	}
	var resp dsrGetDcSiteCoverageWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("DsrGetDcSiteCoverageW: %w", err)
		return
	}
	SiteNames = resp.SiteNames
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("DsrGetDcSiteCoverageW failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

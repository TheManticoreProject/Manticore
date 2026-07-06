package functions

// IDL source: [MS-DNSP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dnsp/83136c8e-f5ea-4ec5-bf33-2134053d33bd
// A fetched copy is kept at ms-dnsp.idl in the interface directory.

import (
	"fmt"

	DnsServer "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/50abc2a4-574d-40b3-9d66-ee4fd5fba076/5.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdnsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dnsp"
)

// r_DnssrvQueryRequest carries the [in] parameters of R_DnssrvQuery.
type r_DnssrvQueryRequest struct {
	PwszServerName *ndr.WSTR `ndr:"unique"`
	PszZone        *ndr.STR  `ndr:"unique"`
	PszOperation   *ndr.STR  `ndr:"unique"`
}

func (*r_DnssrvQueryRequest) Opnum() uint16 { return DnsServer.OpnumR_DnssrvQuery }

// r_DnssrvQueryResponse carries the [out] parameters and return value of R_DnssrvQuery.
type r_DnssrvQueryResponse struct {
	PdwTypeId ndr.DWORD
	PpData    msdnsp.DNSSRV_RPC_UNION
	Status    ndr.DWORD `ndr:"retval"`
}

// R_DnssrvQuery calls R_DnssrvQuery (opnum 1) ([MS-DNSP] — verify the parameter
// modeling and status handling).
func R_DnssrvQuery(rpc ndr.Invoker, pwszServerName *ndr.WSTR, pszZone *ndr.STR, pszOperation *ndr.STR) (PdwTypeId ndr.DWORD, PpData msdnsp.DNSSRV_RPC_UNION, err error) {
	req := &r_DnssrvQueryRequest{
		PwszServerName: pwszServerName,
		PszZone:        pszZone,
		PszOperation:   pszOperation,
	}
	var resp r_DnssrvQueryResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DnssrvQuery: %w", err)
		return
	}
	PdwTypeId = resp.PdwTypeId
	PpData = resp.PpData
	if uint32(resp.Status) != DnsServer.StatusSuccess {
		err = fmt.Errorf("R_DnssrvQuery failed: %s", DnsServer.StatusString(uint32(resp.Status)))
	}
	return
}

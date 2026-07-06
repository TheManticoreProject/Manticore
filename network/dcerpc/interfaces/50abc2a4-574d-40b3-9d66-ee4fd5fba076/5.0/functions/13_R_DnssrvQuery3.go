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

// r_DnssrvQuery3Request carries the [in] parameters of R_DnssrvQuery3.
type r_DnssrvQuery3Request struct {
	DwClientVersion  ndr.DWORD
	DwSettingFlags   ndr.DWORD
	PwszServerName   *ndr.WSTR `ndr:"unique"`
	PszZone          *ndr.STR  `ndr:"unique"`
	PszZoneScopeName *ndr.WSTR `ndr:"unique"`
	PszOperation     *ndr.STR  `ndr:"unique"`
}

func (*r_DnssrvQuery3Request) Opnum() uint16 { return DnsServer.OpnumR_DnssrvQuery3 }

// r_DnssrvQuery3Response carries the [out] parameters and return value of R_DnssrvQuery3.
type r_DnssrvQuery3Response struct {
	PdwTypeId ndr.DWORD
	PpData    msdnsp.DNSSRV_RPC_UNION
	Status    ndr.DWORD `ndr:"retval"`
}

// R_DnssrvQuery3 calls R_DnssrvQuery3 (opnum 13) ([MS-DNSP] — verify the parameter
// modeling and status handling).
func R_DnssrvQuery3(rpc ndr.Invoker, dwClientVersion ndr.DWORD, dwSettingFlags ndr.DWORD, pwszServerName *ndr.WSTR, pszZone *ndr.STR, pszZoneScopeName *ndr.WSTR, pszOperation *ndr.STR) (PdwTypeId ndr.DWORD, PpData msdnsp.DNSSRV_RPC_UNION, err error) {
	req := &r_DnssrvQuery3Request{
		DwClientVersion:  dwClientVersion,
		DwSettingFlags:   dwSettingFlags,
		PwszServerName:   pwszServerName,
		PszZone:          pszZone,
		PszZoneScopeName: pszZoneScopeName,
		PszOperation:     pszOperation,
	}
	var resp r_DnssrvQuery3Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DnssrvQuery3: %w", err)
		return
	}
	PdwTypeId = resp.PdwTypeId
	PpData = resp.PpData
	if uint32(resp.Status) != DnsServer.StatusSuccess {
		err = fmt.Errorf("R_DnssrvQuery3 failed: %s", DnsServer.StatusString(uint32(resp.Status)))
	}
	return
}

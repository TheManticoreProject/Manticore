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

// r_DnssrvQuery4Request carries the [in] parameters of R_DnssrvQuery4.
type r_DnssrvQuery4Request struct {
	DwClientVersion              ndr.DWORD
	DwSettingFlags               ndr.DWORD
	PwszServerName               *ndr.WSTR `ndr:"unique"`
	PwszVirtualizationInstanceID *ndr.WSTR `ndr:"unique"`
	PszZone                      *ndr.STR  `ndr:"unique"`
	PszZoneScopeName             *ndr.WSTR `ndr:"unique"`
	PszOperation                 *ndr.STR  `ndr:"unique"`
}

func (*r_DnssrvQuery4Request) Opnum() uint16 { return DnsServer.OpnumR_DnssrvQuery4 }

// r_DnssrvQuery4Response carries the [out] parameters and return value of R_DnssrvQuery4.
type r_DnssrvQuery4Response struct {
	PdwTypeId ndr.DWORD
	PpData    msdnsp.DNSSRV_RPC_UNION
	Status    ndr.DWORD `ndr:"retval"`
}

// R_DnssrvQuery4 calls R_DnssrvQuery4 (opnum 16) ([MS-DNSP] — verify the parameter
// modeling and status handling).
func R_DnssrvQuery4(rpc ndr.Invoker, dwClientVersion ndr.DWORD, dwSettingFlags ndr.DWORD, pwszServerName *ndr.WSTR, pwszVirtualizationInstanceID *ndr.WSTR, pszZone *ndr.STR, pszZoneScopeName *ndr.WSTR, pszOperation *ndr.STR) (PdwTypeId ndr.DWORD, PpData msdnsp.DNSSRV_RPC_UNION, err error) {
	req := &r_DnssrvQuery4Request{
		DwClientVersion:              dwClientVersion,
		DwSettingFlags:               dwSettingFlags,
		PwszServerName:               pwszServerName,
		PwszVirtualizationInstanceID: pwszVirtualizationInstanceID,
		PszZone:                      pszZone,
		PszZoneScopeName:             pszZoneScopeName,
		PszOperation:                 pszOperation,
	}
	var resp r_DnssrvQuery4Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DnssrvQuery4: %w", err)
		return
	}
	PdwTypeId = resp.PdwTypeId
	PpData = resp.PpData
	if uint32(resp.Status) != DnsServer.StatusSuccess {
		err = fmt.Errorf("R_DnssrvQuery4 failed: %s", DnsServer.StatusString(uint32(resp.Status)))
	}
	return
}

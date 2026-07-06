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

// r_DnssrvComplexOperation3Request carries the [in] parameters of R_DnssrvComplexOperation3.
type r_DnssrvComplexOperation3Request struct {
	DwClientVersion              ndr.DWORD
	DwSettingFlags               ndr.DWORD
	PwszServerName               *ndr.WSTR `ndr:"unique"`
	PwszVirtualizationInstanceID *ndr.WSTR `ndr:"unique"`
	PszZone                      *ndr.STR  `ndr:"unique"`
	PszOperation                 *ndr.STR  `ndr:"unique"`
	DwTypeIn                     ndr.DWORD
	PDataIn                      msdnsp.DNSSRV_RPC_UNION
}

func (*r_DnssrvComplexOperation3Request) Opnum() uint16 {
	return DnsServer.OpnumR_DnssrvComplexOperation3
}

// r_DnssrvComplexOperation3Response carries the [out] parameters and return value of R_DnssrvComplexOperation3.
type r_DnssrvComplexOperation3Response struct {
	PdwTypeOut ndr.DWORD
	PpDataOut  msdnsp.DNSSRV_RPC_UNION
	Status     ndr.DWORD `ndr:"retval"`
}

// R_DnssrvComplexOperation3 calls R_DnssrvComplexOperation3 (opnum 14) ([MS-DNSP] — verify the parameter
// modeling and status handling).
func R_DnssrvComplexOperation3(rpc ndr.Invoker, dwClientVersion ndr.DWORD, dwSettingFlags ndr.DWORD, pwszServerName *ndr.WSTR, pwszVirtualizationInstanceID *ndr.WSTR, pszZone *ndr.STR, pszOperation *ndr.STR, dwTypeIn ndr.DWORD, pDataIn msdnsp.DNSSRV_RPC_UNION) (PdwTypeOut ndr.DWORD, PpDataOut msdnsp.DNSSRV_RPC_UNION, err error) {
	req := &r_DnssrvComplexOperation3Request{
		DwClientVersion:              dwClientVersion,
		DwSettingFlags:               dwSettingFlags,
		PwszServerName:               pwszServerName,
		PwszVirtualizationInstanceID: pwszVirtualizationInstanceID,
		PszZone:                      pszZone,
		PszOperation:                 pszOperation,
		DwTypeIn:                     dwTypeIn,
		PDataIn:                      pDataIn,
	}
	req.PDataIn.Tag = ndr.DWORD(dwTypeIn)
	var resp r_DnssrvComplexOperation3Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DnssrvComplexOperation3: %w", err)
		return
	}
	PdwTypeOut = resp.PdwTypeOut
	PpDataOut = resp.PpDataOut
	if uint32(resp.Status) != DnsServer.StatusSuccess {
		err = fmt.Errorf("R_DnssrvComplexOperation3 failed: %s", DnsServer.StatusString(uint32(resp.Status)))
	}
	return
}

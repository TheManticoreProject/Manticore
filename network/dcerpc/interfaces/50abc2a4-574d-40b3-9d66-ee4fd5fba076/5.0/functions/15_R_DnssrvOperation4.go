package functions

import (
	"fmt"

	DnsServer "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/50abc2a4-574d-40b3-9d66-ee4fd5fba076/5.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdnsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dnsp"
)

// r_DnssrvOperation4Request carries the [in] parameters of R_DnssrvOperation4.
type r_DnssrvOperation4Request struct {
	DwClientVersion              ndr.DWORD
	DwSettingFlags               ndr.DWORD
	PwszServerName               *ndr.WSTR `ndr:"unique"`
	PwszVirtualizationInstanceID *ndr.WSTR `ndr:"unique"`
	PszZone                      *ndr.STR  `ndr:"unique"`
	PwszZoneScopeName            *ndr.WSTR `ndr:"unique"`
	DwContext                    ndr.DWORD
	PszOperation                 *ndr.STR `ndr:"unique"`
	DwTypeId                     ndr.DWORD
	PData                        msdnsp.DNSSRV_RPC_UNION
}

func (*r_DnssrvOperation4Request) Opnum() uint16 { return DnsServer.OpnumR_DnssrvOperation4 }

// r_DnssrvOperation4Response carries the [out] parameters and return value of R_DnssrvOperation4.
type r_DnssrvOperation4Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DnssrvOperation4 calls R_DnssrvOperation4 (opnum 15) ([MS-DNSP] — verify the parameter
// modeling and status handling).
func R_DnssrvOperation4(rpc ndr.Invoker, dwClientVersion ndr.DWORD, dwSettingFlags ndr.DWORD, pwszServerName *ndr.WSTR, pwszVirtualizationInstanceID *ndr.WSTR, pszZone *ndr.STR, pwszZoneScopeName *ndr.WSTR, dwContext ndr.DWORD, pszOperation *ndr.STR, dwTypeId ndr.DWORD, pData msdnsp.DNSSRV_RPC_UNION) (err error) {
	req := &r_DnssrvOperation4Request{
		DwClientVersion:              dwClientVersion,
		DwSettingFlags:               dwSettingFlags,
		PwszServerName:               pwszServerName,
		PwszVirtualizationInstanceID: pwszVirtualizationInstanceID,
		PszZone:                      pszZone,
		PwszZoneScopeName:            pwszZoneScopeName,
		DwContext:                    dwContext,
		PszOperation:                 pszOperation,
		DwTypeId:                     dwTypeId,
		PData:                        pData,
	}
	req.PData.Tag = ndr.DWORD(dwTypeId)
	var resp r_DnssrvOperation4Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DnssrvOperation4: %w", err)
		return
	}
	if uint32(resp.Status) != DnsServer.StatusSuccess {
		err = fmt.Errorf("R_DnssrvOperation4 failed: %s", DnsServer.StatusString(uint32(resp.Status)))
	}
	return
}

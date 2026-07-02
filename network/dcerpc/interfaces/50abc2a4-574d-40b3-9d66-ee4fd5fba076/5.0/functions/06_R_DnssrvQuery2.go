package functions

import (
	"fmt"

	DnsServer "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/50abc2a4-574d-40b3-9d66-ee4fd5fba076/5.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdnsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dnsp"
)

// r_DnssrvQuery2Request carries the [in] parameters of R_DnssrvQuery2.
type r_DnssrvQuery2Request struct {
	DwClientVersion ndr.DWORD
	DwSettingFlags  ndr.DWORD
	PwszServerName  *ndr.WSTR `ndr:"unique"`
	PszZone         *ndr.STR  `ndr:"unique"`
	PszOperation    *ndr.STR  `ndr:"unique"`
}

func (*r_DnssrvQuery2Request) Opnum() uint16 { return DnsServer.OpnumR_DnssrvQuery2 }

// r_DnssrvQuery2Response carries the [out] parameters and return value of R_DnssrvQuery2.
type r_DnssrvQuery2Response struct {
	PdwTypeId ndr.DWORD
	PpData    msdnsp.DNSSRV_RPC_UNION
	Status    ndr.DWORD `ndr:"retval"`
}

// R_DnssrvQuery2 calls R_DnssrvQuery2 (opnum 6) ([MS-DNSP] — verify the parameter
// modeling and status handling).
func R_DnssrvQuery2(rpc ndr.Invoker, dwClientVersion ndr.DWORD, dwSettingFlags ndr.DWORD, pwszServerName *ndr.WSTR, pszZone *ndr.STR, pszOperation *ndr.STR) (PdwTypeId ndr.DWORD, PpData msdnsp.DNSSRV_RPC_UNION, err error) {
	req := &r_DnssrvQuery2Request{
		DwClientVersion: dwClientVersion,
		DwSettingFlags:  dwSettingFlags,
		PwszServerName:  pwszServerName,
		PszZone:         pszZone,
		PszOperation:    pszOperation,
	}
	var resp r_DnssrvQuery2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DnssrvQuery2: %w", err)
		return
	}
	PdwTypeId = resp.PdwTypeId
	PpData = resp.PpData
	if uint32(resp.Status) != DnsServer.StatusSuccess {
		err = fmt.Errorf("R_DnssrvQuery2 failed: %s", DnsServer.StatusString(uint32(resp.Status)))
	}
	return
}

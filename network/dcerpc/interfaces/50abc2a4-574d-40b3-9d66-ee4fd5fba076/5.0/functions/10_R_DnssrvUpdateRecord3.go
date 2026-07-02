package functions

import (
	"fmt"

	DnsServer "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/50abc2a4-574d-40b3-9d66-ee4fd5fba076/5.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdnsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dnsp"
)

// r_DnssrvUpdateRecord3Request carries the [in] parameters of R_DnssrvUpdateRecord3.
type r_DnssrvUpdateRecord3Request struct {
	DwClientVersion ndr.DWORD
	DwSettingFlags  ndr.DWORD
	PwszServerName  *ndr.WSTR `ndr:"unique"`
	PszZone         *ndr.STR  `ndr:"unique"`
	PwszZoneScope   *ndr.WSTR `ndr:"unique"`
	PszNodeName     ndr.STR
	PAddRecord      *msdnsp.DNS_RPC_RECORD `ndr:"unique"`
	PDeleteRecord   *msdnsp.DNS_RPC_RECORD `ndr:"unique"`
}

func (*r_DnssrvUpdateRecord3Request) Opnum() uint16 { return DnsServer.OpnumR_DnssrvUpdateRecord3 }

// r_DnssrvUpdateRecord3Response carries the [out] parameters and return value of R_DnssrvUpdateRecord3.
type r_DnssrvUpdateRecord3Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DnssrvUpdateRecord3 calls R_DnssrvUpdateRecord3 (opnum 10) ([MS-DNSP] — verify the parameter
// modeling and status handling).
func R_DnssrvUpdateRecord3(rpc ndr.Invoker, dwClientVersion ndr.DWORD, dwSettingFlags ndr.DWORD, pwszServerName *ndr.WSTR, pszZone *ndr.STR, pwszZoneScope *ndr.WSTR, pszNodeName ndr.STR, pAddRecord *msdnsp.DNS_RPC_RECORD, pDeleteRecord *msdnsp.DNS_RPC_RECORD) (err error) {
	req := &r_DnssrvUpdateRecord3Request{
		DwClientVersion: dwClientVersion,
		DwSettingFlags:  dwSettingFlags,
		PwszServerName:  pwszServerName,
		PszZone:         pszZone,
		PwszZoneScope:   pwszZoneScope,
		PszNodeName:     pszNodeName,
		PAddRecord:      pAddRecord,
		PDeleteRecord:   pDeleteRecord,
	}
	var resp r_DnssrvUpdateRecord3Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DnssrvUpdateRecord3: %w", err)
		return
	}
	if uint32(resp.Status) != DnsServer.StatusSuccess {
		err = fmt.Errorf("R_DnssrvUpdateRecord3 failed: %s", DnsServer.StatusString(uint32(resp.Status)))
	}
	return
}

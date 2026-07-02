package functions

import (
	"fmt"

	DnsServer "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/50abc2a4-574d-40b3-9d66-ee4fd5fba076/5.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdnsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dnsp"
)

// r_DnssrvUpdateRecord2Request carries the [in] parameters of R_DnssrvUpdateRecord2.
type r_DnssrvUpdateRecord2Request struct {
	DwClientVersion ndr.DWORD
	DwSettingFlags  ndr.DWORD
	PwszServerName  *ndr.WSTR `ndr:"unique"`
	PszZone         *ndr.STR  `ndr:"unique"`
	PszNodeName     ndr.STR
	PAddRecord      *msdnsp.DNS_RPC_RECORD `ndr:"unique"`
	PDeleteRecord   *msdnsp.DNS_RPC_RECORD `ndr:"unique"`
}

func (*r_DnssrvUpdateRecord2Request) Opnum() uint16 { return DnsServer.OpnumR_DnssrvUpdateRecord2 }

// r_DnssrvUpdateRecord2Response carries the [out] parameters and return value of R_DnssrvUpdateRecord2.
type r_DnssrvUpdateRecord2Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DnssrvUpdateRecord2 calls R_DnssrvUpdateRecord2 (opnum 9) ([MS-DNSP] — verify the parameter
// modeling and status handling).
func R_DnssrvUpdateRecord2(rpc ndr.Invoker, dwClientVersion ndr.DWORD, dwSettingFlags ndr.DWORD, pwszServerName *ndr.WSTR, pszZone *ndr.STR, pszNodeName ndr.STR, pAddRecord *msdnsp.DNS_RPC_RECORD, pDeleteRecord *msdnsp.DNS_RPC_RECORD) (err error) {
	req := &r_DnssrvUpdateRecord2Request{
		DwClientVersion: dwClientVersion,
		DwSettingFlags:  dwSettingFlags,
		PwszServerName:  pwszServerName,
		PszZone:         pszZone,
		PszNodeName:     pszNodeName,
		PAddRecord:      pAddRecord,
		PDeleteRecord:   pDeleteRecord,
	}
	var resp r_DnssrvUpdateRecord2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DnssrvUpdateRecord2: %w", err)
		return
	}
	if uint32(resp.Status) != DnsServer.StatusSuccess {
		err = fmt.Errorf("R_DnssrvUpdateRecord2 failed: %s", DnsServer.StatusString(uint32(resp.Status)))
	}
	return
}

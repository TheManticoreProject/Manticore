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

// r_DnssrvUpdateRecordRequest carries the [in] parameters of R_DnssrvUpdateRecord.
type r_DnssrvUpdateRecordRequest struct {
	PwszServerName *ndr.WSTR `ndr:"unique"`
	PszZone        *ndr.STR  `ndr:"unique"`
	PszNodeName    ndr.STR
	PAddRecord     *msdnsp.DNS_RPC_RECORD `ndr:"unique"`
	PDeleteRecord  *msdnsp.DNS_RPC_RECORD `ndr:"unique"`
}

func (*r_DnssrvUpdateRecordRequest) Opnum() uint16 { return DnsServer.OpnumR_DnssrvUpdateRecord }

// r_DnssrvUpdateRecordResponse carries the [out] parameters and return value of R_DnssrvUpdateRecord.
type r_DnssrvUpdateRecordResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DnssrvUpdateRecord calls R_DnssrvUpdateRecord (opnum 4) ([MS-DNSP] — verify the parameter
// modeling and status handling).
func R_DnssrvUpdateRecord(rpc ndr.Invoker, pwszServerName *ndr.WSTR, pszZone *ndr.STR, pszNodeName ndr.STR, pAddRecord *msdnsp.DNS_RPC_RECORD, pDeleteRecord *msdnsp.DNS_RPC_RECORD) (err error) {
	req := &r_DnssrvUpdateRecordRequest{
		PwszServerName: pwszServerName,
		PszZone:        pszZone,
		PszNodeName:    pszNodeName,
		PAddRecord:     pAddRecord,
		PDeleteRecord:  pDeleteRecord,
	}
	var resp r_DnssrvUpdateRecordResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DnssrvUpdateRecord: %w", err)
		return
	}
	if uint32(resp.Status) != DnsServer.StatusSuccess {
		err = fmt.Errorf("R_DnssrvUpdateRecord failed: %s", DnsServer.StatusString(uint32(resp.Status)))
	}
	return
}

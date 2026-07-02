package functions

import (
	"fmt"

	DnsServer "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/50abc2a4-574d-40b3-9d66-ee4fd5fba076/5.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdnsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dnsp"
)

// r_DnssrvOperationRequest carries the [in] parameters of R_DnssrvOperation.
type r_DnssrvOperationRequest struct {
	PwszServerName *ndr.WSTR `ndr:"unique"`
	PszZone        *ndr.STR  `ndr:"unique"`
	DwContext      ndr.DWORD
	PszOperation   *ndr.STR `ndr:"unique"`
	DwTypeId       ndr.DWORD
	PData          msdnsp.DNSSRV_RPC_UNION
}

func (*r_DnssrvOperationRequest) Opnum() uint16 { return DnsServer.OpnumR_DnssrvOperation }

// r_DnssrvOperationResponse carries the [out] parameters and return value of R_DnssrvOperation.
type r_DnssrvOperationResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DnssrvOperation calls R_DnssrvOperation (opnum 0) ([MS-DNSP] — verify the parameter
// modeling and status handling).
func R_DnssrvOperation(rpc ndr.Invoker, pwszServerName *ndr.WSTR, pszZone *ndr.STR, dwContext ndr.DWORD, pszOperation *ndr.STR, dwTypeId ndr.DWORD, pData msdnsp.DNSSRV_RPC_UNION) (err error) {
	req := &r_DnssrvOperationRequest{
		PwszServerName: pwszServerName,
		PszZone:        pszZone,
		DwContext:      dwContext,
		PszOperation:   pszOperation,
		DwTypeId:       dwTypeId,
		PData:          pData,
	}
	req.PData.Tag = ndr.DWORD(dwTypeId)
	var resp r_DnssrvOperationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DnssrvOperation: %w", err)
		return
	}
	if uint32(resp.Status) != DnsServer.StatusSuccess {
		err = fmt.Errorf("R_DnssrvOperation failed: %s", DnsServer.StatusString(uint32(resp.Status)))
	}
	return
}

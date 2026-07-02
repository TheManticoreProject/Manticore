package functions

import (
	"fmt"

	DnsServer "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/50abc2a4-574d-40b3-9d66-ee4fd5fba076/5.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdnsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dnsp"
)

// r_DnssrvComplexOperationRequest carries the [in] parameters of R_DnssrvComplexOperation.
type r_DnssrvComplexOperationRequest struct {
	PwszServerName *ndr.WSTR `ndr:"unique"`
	PszZone        *ndr.STR  `ndr:"unique"`
	PszOperation   *ndr.STR  `ndr:"unique"`
	DwTypeIn       ndr.DWORD
	PDataIn        msdnsp.DNSSRV_RPC_UNION
}

func (*r_DnssrvComplexOperationRequest) Opnum() uint16 {
	return DnsServer.OpnumR_DnssrvComplexOperation
}

// r_DnssrvComplexOperationResponse carries the [out] parameters and return value of R_DnssrvComplexOperation.
type r_DnssrvComplexOperationResponse struct {
	PdwTypeOut ndr.DWORD
	PpDataOut  msdnsp.DNSSRV_RPC_UNION
	Status     ndr.DWORD `ndr:"retval"`
}

// R_DnssrvComplexOperation calls R_DnssrvComplexOperation (opnum 2) ([MS-DNSP] — verify the parameter
// modeling and status handling).
func R_DnssrvComplexOperation(rpc ndr.Invoker, pwszServerName *ndr.WSTR, pszZone *ndr.STR, pszOperation *ndr.STR, dwTypeIn ndr.DWORD, pDataIn msdnsp.DNSSRV_RPC_UNION) (PdwTypeOut ndr.DWORD, PpDataOut msdnsp.DNSSRV_RPC_UNION, err error) {
	req := &r_DnssrvComplexOperationRequest{
		PwszServerName: pwszServerName,
		PszZone:        pszZone,
		PszOperation:   pszOperation,
		DwTypeIn:       dwTypeIn,
		PDataIn:        pDataIn,
	}
	req.PDataIn.Tag = ndr.DWORD(dwTypeIn)
	var resp r_DnssrvComplexOperationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DnssrvComplexOperation: %w", err)
		return
	}
	PdwTypeOut = resp.PdwTypeOut
	PpDataOut = resp.PpDataOut
	if uint32(resp.Status) != DnsServer.StatusSuccess {
		err = fmt.Errorf("R_DnssrvComplexOperation failed: %s", DnsServer.StatusString(uint32(resp.Status)))
	}
	return
}

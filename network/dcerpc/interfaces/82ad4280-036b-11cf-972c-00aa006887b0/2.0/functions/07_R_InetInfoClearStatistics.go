package functions

// IDL source: [MS-IRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-irp/ed7e5940-9700-4a1f-8555-de29f99fe115
// A fetched copy is kept at ms-irp.idl in the interface directory.

import (
	"fmt"

	inetinfo "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82ad4280-036b-11cf-972c-00aa006887b0/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_InetInfoClearStatisticsRequest carries the [in] parameters of R_InetInfoClearStatistics.
type r_InetInfoClearStatisticsRequest struct {
	PszServer    *ndr.WSTR `ndr:"unique"`
	DwServerMask ndr.DWORD
}

func (*r_InetInfoClearStatisticsRequest) Opnum() uint16 {
	return inetinfo.OpnumR_InetInfoClearStatistics
}

// r_InetInfoClearStatisticsResponse carries the [out] parameters and return value of R_InetInfoClearStatistics.
type r_InetInfoClearStatisticsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_InetInfoClearStatistics calls R_InetInfoClearStatistics (opnum 7) ([MS-IRP] — verify the parameter
// modeling and status handling).
func R_InetInfoClearStatistics(rpc ndr.Invoker, pszServer *ndr.WSTR, dwServerMask ndr.DWORD) (err error) {
	req := &r_InetInfoClearStatisticsRequest{
		PszServer:    pszServer,
		DwServerMask: dwServerMask,
	}
	var resp r_InetInfoClearStatisticsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_InetInfoClearStatistics: %w", err)
		return
	}
	if uint32(resp.Status) != inetinfo.ErrorSuccess {
		err = fmt.Errorf("R_InetInfoClearStatistics failed: %s", inetinfo.StatusString(uint32(resp.Status)))
	}
	return
}

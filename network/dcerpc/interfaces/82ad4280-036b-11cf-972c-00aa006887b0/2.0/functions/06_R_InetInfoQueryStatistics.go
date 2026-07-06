package functions

// IDL source: [MS-IRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-irp/ed7e5940-9700-4a1f-8555-de29f99fe115
// A fetched copy is kept at ms-irp.idl in the interface directory.

import (
	"fmt"

	inetinfo "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82ad4280-036b-11cf-972c-00aa006887b0/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msirp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-irp"
)

// r_InetInfoQueryStatisticsRequest carries the [in] parameters of R_InetInfoQueryStatistics.
type r_InetInfoQueryStatisticsRequest struct {
	PszServer    *ndr.WSTR `ndr:"unique"`
	Level        ndr.DWORD
	DwServerMask ndr.DWORD
}

func (*r_InetInfoQueryStatisticsRequest) Opnum() uint16 {
	return inetinfo.OpnumR_InetInfoQueryStatistics
}

// r_InetInfoQueryStatisticsResponse carries the [out] parameters and return value of R_InetInfoQueryStatistics.
type r_InetInfoQueryStatisticsResponse struct {
	StatsInfo msirp.INET_INFO_STATISTICS_INFO
	Status    ndr.DWORD `ndr:"retval"`
}

// R_InetInfoQueryStatistics calls R_InetInfoQueryStatistics (opnum 6) ([MS-IRP] — verify the parameter
// modeling and status handling).
func R_InetInfoQueryStatistics(rpc ndr.Invoker, pszServer *ndr.WSTR, level ndr.DWORD, dwServerMask ndr.DWORD) (StatsInfo msirp.INET_INFO_STATISTICS_INFO, err error) {
	req := &r_InetInfoQueryStatisticsRequest{
		PszServer:    pszServer,
		Level:        level,
		DwServerMask: dwServerMask,
	}
	var resp r_InetInfoQueryStatisticsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_InetInfoQueryStatistics: %w", err)
		return
	}
	StatsInfo = resp.StatsInfo
	if uint32(resp.Status) != inetinfo.ErrorSuccess {
		err = fmt.Errorf("R_InetInfoQueryStatistics failed: %s", inetinfo.StatusString(uint32(resp.Status)))
	}
	return
}

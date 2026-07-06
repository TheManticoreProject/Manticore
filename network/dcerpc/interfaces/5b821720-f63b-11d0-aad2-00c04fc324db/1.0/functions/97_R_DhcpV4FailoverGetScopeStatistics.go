package functions

// IDL source: [MS-DHCPM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dhcpm/d1932d08-3249-44cb-90f1-8661f8fb5b90
// A fetched copy is kept at ms-dhcpm.idl in the interface directory.

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4FailoverGetScopeStatisticsRequest carries the [in] parameters of R_DhcpV4FailoverGetScopeStatistics.
type r_DhcpV4FailoverGetScopeStatisticsRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ScopeId         ndr.DWORD
}

func (*r_DhcpV4FailoverGetScopeStatisticsRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4FailoverGetScopeStatistics
}

// r_DhcpV4FailoverGetScopeStatisticsResponse carries the [out] parameters and return value of R_DhcpV4FailoverGetScopeStatistics.
type r_DhcpV4FailoverGetScopeStatisticsResponse struct {
	PStats *msdhcpm.DHCP_FAILOVER_STATISTICS `ndr:"unique"`
	Status ndr.DWORD                         `ndr:"retval"`
}

// R_DhcpV4FailoverGetScopeStatistics calls R_DhcpV4FailoverGetScopeStatistics (opnum 97) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4FailoverGetScopeStatistics(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, scopeId ndr.DWORD) (PStats *msdhcpm.DHCP_FAILOVER_STATISTICS, err error) {
	req := &r_DhcpV4FailoverGetScopeStatisticsRequest{
		ServerIpAddress: serverIpAddress,
		ScopeId:         scopeId,
	}
	var resp r_DhcpV4FailoverGetScopeStatisticsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4FailoverGetScopeStatistics: %w", err)
		return
	}
	PStats = resp.PStats
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4FailoverGetScopeStatistics failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

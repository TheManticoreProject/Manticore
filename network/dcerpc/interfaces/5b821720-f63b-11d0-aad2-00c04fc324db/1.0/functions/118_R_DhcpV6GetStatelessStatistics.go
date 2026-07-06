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

// r_DhcpV6GetStatelessStatisticsRequest carries the [in] parameters of R_DhcpV6GetStatelessStatistics.
type r_DhcpV6GetStatelessStatisticsRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpV6GetStatelessStatisticsRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV6GetStatelessStatistics
}

// r_DhcpV6GetStatelessStatisticsResponse carries the [out] parameters and return value of R_DhcpV6GetStatelessStatistics.
type r_DhcpV6GetStatelessStatisticsResponse struct {
	StatelessStats *msdhcpm.DHCPV6_STATELESS_STATS `ndr:"unique"`
	Status         ndr.DWORD                       `ndr:"retval"`
}

// R_DhcpV6GetStatelessStatistics calls R_DhcpV6GetStatelessStatistics (opnum 118) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV6GetStatelessStatistics(rpc ndr.Invoker, serverIpAddress *ndr.WSTR) (StatelessStats *msdhcpm.DHCPV6_STATELESS_STATS, err error) {
	req := &r_DhcpV6GetStatelessStatisticsRequest{
		ServerIpAddress: serverIpAddress,
	}
	var resp r_DhcpV6GetStatelessStatisticsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV6GetStatelessStatistics: %w", err)
		return
	}
	StatelessStats = resp.StatelessStats
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV6GetStatelessStatistics failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

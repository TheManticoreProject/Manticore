package functions

// IDL source: [MS-DHCPM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dhcpm/d1932d08-3249-44cb-90f1-8661f8fb5b90
// A fetched copy is kept at ms-dhcpm.idl in the interface directory.

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpV4FailoverGetSystemTimeRequest carries the [in] parameters of R_DhcpV4FailoverGetSystemTime.
type r_DhcpV4FailoverGetSystemTimeRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpV4FailoverGetSystemTimeRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4FailoverGetSystemTime
}

// r_DhcpV4FailoverGetSystemTimeResponse carries the [out] parameters and return value of R_DhcpV4FailoverGetSystemTime.
type r_DhcpV4FailoverGetSystemTimeResponse struct {
	PTime  ndr.DWORD
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4FailoverGetSystemTime calls R_DhcpV4FailoverGetSystemTime (opnum 99) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4FailoverGetSystemTime(rpc ndr.Invoker, serverIpAddress *ndr.WSTR) (PTime ndr.DWORD, err error) {
	req := &r_DhcpV4FailoverGetSystemTimeRequest{
		ServerIpAddress: serverIpAddress,
	}
	var resp r_DhcpV4FailoverGetSystemTimeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4FailoverGetSystemTime: %w", err)
		return
	}
	PTime = resp.PTime
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4FailoverGetSystemTime failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

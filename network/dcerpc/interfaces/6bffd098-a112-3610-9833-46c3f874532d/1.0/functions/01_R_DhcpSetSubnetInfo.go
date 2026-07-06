package functions

// IDL source: [MS-DHCPM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dhcpm/d1932d08-3249-44cb-90f1-8661f8fb5b90
// A fetched copy is kept at ms-dhcpm.idl in the interface directory.

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpSetSubnetInfoRequest carries the [in] parameters of R_DhcpSetSubnetInfo.
type r_DhcpSetSubnetInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   ndr.DWORD
	SubnetInfo      msdhcpm.DHCP_SUBNET_INFO
}

func (*r_DhcpSetSubnetInfoRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpSetSubnetInfo }

// r_DhcpSetSubnetInfoResponse carries the [out] parameters and return value of R_DhcpSetSubnetInfo.
type r_DhcpSetSubnetInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetSubnetInfo calls R_DhcpSetSubnetInfo (opnum 1) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetSubnetInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, subnetInfo msdhcpm.DHCP_SUBNET_INFO) (err error) {
	req := &r_DhcpSetSubnetInfoRequest{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
		SubnetInfo:      subnetInfo,
	}
	var resp r_DhcpSetSubnetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetSubnetInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetSubnetInfo failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}

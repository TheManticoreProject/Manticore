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

// r_DhcpCreateSubnetRequest carries the [in] parameters of R_DhcpCreateSubnet.
type r_DhcpCreateSubnetRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   ndr.DWORD
	SubnetInfo      msdhcpm.DHCP_SUBNET_INFO
}

func (*r_DhcpCreateSubnetRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpCreateSubnet }

// r_DhcpCreateSubnetResponse carries the [out] parameters and return value of R_DhcpCreateSubnet.
type r_DhcpCreateSubnetResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpCreateSubnet calls R_DhcpCreateSubnet (opnum 0) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpCreateSubnet(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, subnetInfo msdhcpm.DHCP_SUBNET_INFO) (err error) {
	req := &r_DhcpCreateSubnetRequest{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
		SubnetInfo:      subnetInfo,
	}
	var resp r_DhcpCreateSubnetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpCreateSubnet: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpCreateSubnet failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}

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

// r_DhcpAddSubnetElementV5Request carries the [in] parameters of R_DhcpAddSubnetElementV5.
type r_DhcpAddSubnetElementV5Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   ndr.DWORD
	AddElementInfo  msdhcpm.DHCP_SUBNET_ELEMENT_DATA_V5
}

func (*r_DhcpAddSubnetElementV5Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpAddSubnetElementV5 }

// r_DhcpAddSubnetElementV5Response carries the [out] parameters and return value of R_DhcpAddSubnetElementV5.
type r_DhcpAddSubnetElementV5Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpAddSubnetElementV5 calls R_DhcpAddSubnetElementV5 (opnum 37) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpAddSubnetElementV5(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, addElementInfo msdhcpm.DHCP_SUBNET_ELEMENT_DATA_V5) (err error) {
	req := &r_DhcpAddSubnetElementV5Request{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
		AddElementInfo:  addElementInfo,
	}
	var resp r_DhcpAddSubnetElementV5Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpAddSubnetElementV5: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpAddSubnetElementV5 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

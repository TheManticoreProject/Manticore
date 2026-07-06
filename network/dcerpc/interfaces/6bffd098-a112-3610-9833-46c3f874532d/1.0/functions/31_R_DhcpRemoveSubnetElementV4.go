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

// r_DhcpRemoveSubnetElementV4Request carries the [in] parameters of R_DhcpRemoveSubnetElementV4.
type r_DhcpRemoveSubnetElementV4Request struct {
	ServerIpAddress   *ndr.WSTR `ndr:"unique"`
	SubnetAddress     ndr.DWORD
	RemoveElementInfo msdhcpm.DHCP_SUBNET_ELEMENT_DATA_V4
	ForceFlag         msdhcpm.DHCP_FORCE_FLAG
}

func (*r_DhcpRemoveSubnetElementV4Request) Opnum() uint16 {
	return dhcpsrv.OpnumR_DhcpRemoveSubnetElementV4
}

// r_DhcpRemoveSubnetElementV4Response carries the [out] parameters and return value of R_DhcpRemoveSubnetElementV4.
type r_DhcpRemoveSubnetElementV4Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpRemoveSubnetElementV4 calls R_DhcpRemoveSubnetElementV4 (opnum 31) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpRemoveSubnetElementV4(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, removeElementInfo msdhcpm.DHCP_SUBNET_ELEMENT_DATA_V4, forceFlag msdhcpm.DHCP_FORCE_FLAG) (err error) {
	req := &r_DhcpRemoveSubnetElementV4Request{
		ServerIpAddress:   serverIpAddress,
		SubnetAddress:     subnetAddress,
		RemoveElementInfo: removeElementInfo,
		ForceFlag:         forceFlag,
	}
	var resp r_DhcpRemoveSubnetElementV4Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpRemoveSubnetElementV4: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpRemoveSubnetElementV4 failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4RemovePolicyRangeRequest carries the [in] parameters of R_DhcpV4RemovePolicyRange.
type r_DhcpV4RemovePolicyRangeRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   ndr.DWORD
	PolicyName      *ndr.WSTR `ndr:"unique"`
	Range           msdhcpm.DHCP_IP_RANGE
}

func (*r_DhcpV4RemovePolicyRangeRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4RemovePolicyRange
}

// r_DhcpV4RemovePolicyRangeResponse carries the [out] parameters and return value of R_DhcpV4RemovePolicyRange.
type r_DhcpV4RemovePolicyRangeResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4RemovePolicyRange calls R_DhcpV4RemovePolicyRange (opnum 114) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4RemovePolicyRange(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, policyName *ndr.WSTR, range_ msdhcpm.DHCP_IP_RANGE) (err error) {
	req := &r_DhcpV4RemovePolicyRangeRequest{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
		PolicyName:      policyName,
		Range:           range_,
	}
	var resp r_DhcpV4RemovePolicyRangeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4RemovePolicyRange: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4RemovePolicyRange failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

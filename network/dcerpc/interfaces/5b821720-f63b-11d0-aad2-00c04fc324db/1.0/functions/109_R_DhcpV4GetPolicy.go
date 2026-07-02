package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4GetPolicyRequest carries the [in] parameters of R_DhcpV4GetPolicy.
type r_DhcpV4GetPolicyRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ServerPolicy    ndr.BOOL
	SubnetAddress   ndr.DWORD
	PolicyName      *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpV4GetPolicyRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpV4GetPolicy }

// r_DhcpV4GetPolicyResponse carries the [out] parameters and return value of R_DhcpV4GetPolicy.
type r_DhcpV4GetPolicyResponse struct {
	Policy *msdhcpm.DHCP_POLICY `ndr:"unique"`
	Status ndr.DWORD            `ndr:"retval"`
}

// R_DhcpV4GetPolicy calls R_DhcpV4GetPolicy (opnum 109) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4GetPolicy(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, serverPolicy ndr.BOOL, subnetAddress ndr.DWORD, policyName *ndr.WSTR) (Policy *msdhcpm.DHCP_POLICY, err error) {
	req := &r_DhcpV4GetPolicyRequest{
		ServerIpAddress: serverIpAddress,
		ServerPolicy:    serverPolicy,
		SubnetAddress:   subnetAddress,
		PolicyName:      policyName,
	}
	var resp r_DhcpV4GetPolicyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4GetPolicy: %w", err)
		return
	}
	Policy = resp.Policy
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4GetPolicy failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

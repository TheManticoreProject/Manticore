package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpV4DeletePolicyRequest carries the [in] parameters of R_DhcpV4DeletePolicy.
type r_DhcpV4DeletePolicyRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ServerPolicy    ndr.BOOL
	SubnetAddress   ndr.DWORD
	PolicyName      *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpV4DeletePolicyRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpV4DeletePolicy }

// r_DhcpV4DeletePolicyResponse carries the [out] parameters and return value of R_DhcpV4DeletePolicy.
type r_DhcpV4DeletePolicyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4DeletePolicy calls R_DhcpV4DeletePolicy (opnum 111) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4DeletePolicy(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, serverPolicy ndr.BOOL, subnetAddress ndr.DWORD, policyName *ndr.WSTR) (err error) {
	req := &r_DhcpV4DeletePolicyRequest{
		ServerIpAddress: serverIpAddress,
		ServerPolicy:    serverPolicy,
		SubnetAddress:   subnetAddress,
		PolicyName:      policyName,
	}
	var resp r_DhcpV4DeletePolicyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4DeletePolicy: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4DeletePolicy failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

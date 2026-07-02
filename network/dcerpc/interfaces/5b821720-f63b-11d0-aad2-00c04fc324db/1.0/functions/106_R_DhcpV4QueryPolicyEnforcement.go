package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpV4QueryPolicyEnforcementRequest carries the [in] parameters of R_DhcpV4QueryPolicyEnforcement.
type r_DhcpV4QueryPolicyEnforcementRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ServerPolicy    ndr.BOOL
	SubnetAddress   ndr.DWORD
}

func (*r_DhcpV4QueryPolicyEnforcementRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4QueryPolicyEnforcement
}

// r_DhcpV4QueryPolicyEnforcementResponse carries the [out] parameters and return value of R_DhcpV4QueryPolicyEnforcement.
type r_DhcpV4QueryPolicyEnforcementResponse struct {
	Enabled ndr.BOOL
	Status  ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4QueryPolicyEnforcement calls R_DhcpV4QueryPolicyEnforcement (opnum 106) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4QueryPolicyEnforcement(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, serverPolicy ndr.BOOL, subnetAddress ndr.DWORD) (Enabled ndr.BOOL, err error) {
	req := &r_DhcpV4QueryPolicyEnforcementRequest{
		ServerIpAddress: serverIpAddress,
		ServerPolicy:    serverPolicy,
		SubnetAddress:   subnetAddress,
	}
	var resp r_DhcpV4QueryPolicyEnforcementResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4QueryPolicyEnforcement: %w", err)
		return
	}
	Enabled = resp.Enabled
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4QueryPolicyEnforcement failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

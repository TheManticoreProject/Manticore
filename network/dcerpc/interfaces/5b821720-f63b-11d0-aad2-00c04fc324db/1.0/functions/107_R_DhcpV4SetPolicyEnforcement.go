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

// r_DhcpV4SetPolicyEnforcementRequest carries the [in] parameters of R_DhcpV4SetPolicyEnforcement.
type r_DhcpV4SetPolicyEnforcementRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ServerPolicy    ndr.BOOL
	SubnetAddress   ndr.DWORD
	Enable          ndr.BOOL
}

func (*r_DhcpV4SetPolicyEnforcementRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4SetPolicyEnforcement
}

// r_DhcpV4SetPolicyEnforcementResponse carries the [out] parameters and return value of R_DhcpV4SetPolicyEnforcement.
type r_DhcpV4SetPolicyEnforcementResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4SetPolicyEnforcement calls R_DhcpV4SetPolicyEnforcement (opnum 107) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4SetPolicyEnforcement(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, serverPolicy ndr.BOOL, subnetAddress ndr.DWORD, enable ndr.BOOL) (err error) {
	req := &r_DhcpV4SetPolicyEnforcementRequest{
		ServerIpAddress: serverIpAddress,
		ServerPolicy:    serverPolicy,
		SubnetAddress:   subnetAddress,
		Enable:          enable,
	}
	var resp r_DhcpV4SetPolicyEnforcementResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4SetPolicyEnforcement: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4SetPolicyEnforcement failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

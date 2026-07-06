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

// r_DhcpV4GetPolicyExRequest carries the [in] parameters of R_DhcpV4GetPolicyEx.
type r_DhcpV4GetPolicyExRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ServerPolicy    ndr.BOOL
	SubnetAddress   ndr.DWORD
	PolicyName      *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpV4GetPolicyExRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpV4GetPolicyEx }

// r_DhcpV4GetPolicyExResponse carries the [out] parameters and return value of R_DhcpV4GetPolicyEx.
type r_DhcpV4GetPolicyExResponse struct {
	Policy *msdhcpm.DHCP_POLICY_EX `ndr:"unique"`
	Status ndr.DWORD               `ndr:"retval"`
}

// R_DhcpV4GetPolicyEx calls R_DhcpV4GetPolicyEx (opnum 127) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4GetPolicyEx(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, serverPolicy ndr.BOOL, subnetAddress ndr.DWORD, policyName *ndr.WSTR) (Policy *msdhcpm.DHCP_POLICY_EX, err error) {
	req := &r_DhcpV4GetPolicyExRequest{
		ServerIpAddress: serverIpAddress,
		ServerPolicy:    serverPolicy,
		SubnetAddress:   subnetAddress,
		PolicyName:      policyName,
	}
	var resp r_DhcpV4GetPolicyExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4GetPolicyEx: %w", err)
		return
	}
	Policy = resp.Policy
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4GetPolicyEx failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

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

// r_DhcpV4SetPolicyRequest carries the [in] parameters of R_DhcpV4SetPolicy.
type r_DhcpV4SetPolicyRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	FieldsModified  ndr.DWORD
	ServerPolicy    ndr.BOOL
	SubnetAddress   ndr.DWORD
	PolicyName      *ndr.WSTR `ndr:"unique"`
	Policy          msdhcpm.DHCP_POLICY
}

func (*r_DhcpV4SetPolicyRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpV4SetPolicy }

// r_DhcpV4SetPolicyResponse carries the [out] parameters and return value of R_DhcpV4SetPolicy.
type r_DhcpV4SetPolicyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4SetPolicy calls R_DhcpV4SetPolicy (opnum 110) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4SetPolicy(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, fieldsModified ndr.DWORD, serverPolicy ndr.BOOL, subnetAddress ndr.DWORD, policyName *ndr.WSTR, policy msdhcpm.DHCP_POLICY) (err error) {
	req := &r_DhcpV4SetPolicyRequest{
		ServerIpAddress: serverIpAddress,
		FieldsModified:  fieldsModified,
		ServerPolicy:    serverPolicy,
		SubnetAddress:   subnetAddress,
		PolicyName:      policyName,
		Policy:          policy,
	}
	var resp r_DhcpV4SetPolicyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4SetPolicy: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4SetPolicy failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

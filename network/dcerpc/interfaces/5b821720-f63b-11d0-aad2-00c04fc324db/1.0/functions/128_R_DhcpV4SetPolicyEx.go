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

// r_DhcpV4SetPolicyExRequest carries the [in] parameters of R_DhcpV4SetPolicyEx.
type r_DhcpV4SetPolicyExRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	FieldsModified  ndr.DWORD
	ServerPolicy    ndr.BOOL
	SubnetAddress   ndr.DWORD
	PolicyName      *ndr.WSTR `ndr:"unique"`
	Policy          msdhcpm.DHCP_POLICY_EX
}

func (*r_DhcpV4SetPolicyExRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpV4SetPolicyEx }

// r_DhcpV4SetPolicyExResponse carries the [out] parameters and return value of R_DhcpV4SetPolicyEx.
type r_DhcpV4SetPolicyExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4SetPolicyEx calls R_DhcpV4SetPolicyEx (opnum 128) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4SetPolicyEx(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, fieldsModified ndr.DWORD, serverPolicy ndr.BOOL, subnetAddress ndr.DWORD, policyName *ndr.WSTR, policy msdhcpm.DHCP_POLICY_EX) (err error) {
	req := &r_DhcpV4SetPolicyExRequest{
		ServerIpAddress: serverIpAddress,
		FieldsModified:  fieldsModified,
		ServerPolicy:    serverPolicy,
		SubnetAddress:   subnetAddress,
		PolicyName:      policyName,
		Policy:          policy,
	}
	var resp r_DhcpV4SetPolicyExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4SetPolicyEx: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4SetPolicyEx failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

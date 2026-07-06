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

// r_DhcpV4FailoverGetRelationshipRequest carries the [in] parameters of R_DhcpV4FailoverGetRelationship.
type r_DhcpV4FailoverGetRelationshipRequest struct {
	ServerIpAddress   *ndr.WSTR `ndr:"unique"`
	PRelationshipName *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpV4FailoverGetRelationshipRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4FailoverGetRelationship
}

// r_DhcpV4FailoverGetRelationshipResponse carries the [out] parameters and return value of R_DhcpV4FailoverGetRelationship.
type r_DhcpV4FailoverGetRelationshipResponse struct {
	PRelationship *msdhcpm.DHCP_FAILOVER_RELATIONSHIP `ndr:"unique"`
	Status        ndr.DWORD                           `ndr:"retval"`
}

// R_DhcpV4FailoverGetRelationship calls R_DhcpV4FailoverGetRelationship (opnum 92) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4FailoverGetRelationship(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, pRelationshipName *ndr.WSTR) (PRelationship *msdhcpm.DHCP_FAILOVER_RELATIONSHIP, err error) {
	req := &r_DhcpV4FailoverGetRelationshipRequest{
		ServerIpAddress:   serverIpAddress,
		PRelationshipName: pRelationshipName,
	}
	var resp r_DhcpV4FailoverGetRelationshipResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4FailoverGetRelationship: %w", err)
		return
	}
	PRelationship = resp.PRelationship
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4FailoverGetRelationship failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

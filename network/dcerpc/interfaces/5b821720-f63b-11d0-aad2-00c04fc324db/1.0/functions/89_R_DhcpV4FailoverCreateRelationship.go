package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4FailoverCreateRelationshipRequest carries the [in] parameters of R_DhcpV4FailoverCreateRelationship.
type r_DhcpV4FailoverCreateRelationshipRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	PRelationship   msdhcpm.DHCP_FAILOVER_RELATIONSHIP
}

func (*r_DhcpV4FailoverCreateRelationshipRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4FailoverCreateRelationship
}

// r_DhcpV4FailoverCreateRelationshipResponse carries the [out] parameters and return value of R_DhcpV4FailoverCreateRelationship.
type r_DhcpV4FailoverCreateRelationshipResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4FailoverCreateRelationship calls R_DhcpV4FailoverCreateRelationship (opnum 89) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4FailoverCreateRelationship(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, pRelationship msdhcpm.DHCP_FAILOVER_RELATIONSHIP) (err error) {
	req := &r_DhcpV4FailoverCreateRelationshipRequest{
		ServerIpAddress: serverIpAddress,
		PRelationship:   pRelationship,
	}
	var resp r_DhcpV4FailoverCreateRelationshipResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4FailoverCreateRelationship: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4FailoverCreateRelationship failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

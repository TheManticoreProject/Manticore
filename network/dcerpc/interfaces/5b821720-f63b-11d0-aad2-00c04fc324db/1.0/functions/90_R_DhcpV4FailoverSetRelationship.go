package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4FailoverSetRelationshipRequest carries the [in] parameters of R_DhcpV4FailoverSetRelationship.
type r_DhcpV4FailoverSetRelationshipRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	PRelationship   msdhcpm.DHCP_FAILOVER_RELATIONSHIP
}

func (*r_DhcpV4FailoverSetRelationshipRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4FailoverSetRelationship
}

// r_DhcpV4FailoverSetRelationshipResponse carries the [out] parameters and return value of R_DhcpV4FailoverSetRelationship.
type r_DhcpV4FailoverSetRelationshipResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4FailoverSetRelationship calls R_DhcpV4FailoverSetRelationship (opnum 90) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4FailoverSetRelationship(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, pRelationship msdhcpm.DHCP_FAILOVER_RELATIONSHIP) (err error) {
	req := &r_DhcpV4FailoverSetRelationshipRequest{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		PRelationship:   pRelationship,
	}
	var resp r_DhcpV4FailoverSetRelationshipResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4FailoverSetRelationship: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4FailoverSetRelationship failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

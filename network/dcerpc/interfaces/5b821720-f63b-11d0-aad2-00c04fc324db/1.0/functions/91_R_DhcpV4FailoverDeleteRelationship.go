package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpV4FailoverDeleteRelationshipRequest carries the [in] parameters of R_DhcpV4FailoverDeleteRelationship.
type r_DhcpV4FailoverDeleteRelationshipRequest struct {
	ServerIpAddress   *ndr.WSTR `ndr:"unique"`
	PRelationshipName *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpV4FailoverDeleteRelationshipRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4FailoverDeleteRelationship
}

// r_DhcpV4FailoverDeleteRelationshipResponse carries the [out] parameters and return value of R_DhcpV4FailoverDeleteRelationship.
type r_DhcpV4FailoverDeleteRelationshipResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4FailoverDeleteRelationship calls R_DhcpV4FailoverDeleteRelationship (opnum 91) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4FailoverDeleteRelationship(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, pRelationshipName *ndr.WSTR) (err error) {
	req := &r_DhcpV4FailoverDeleteRelationshipRequest{
		ServerIpAddress:   serverIpAddress,
		PRelationshipName: pRelationshipName,
	}
	var resp r_DhcpV4FailoverDeleteRelationshipResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4FailoverDeleteRelationship: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4FailoverDeleteRelationship failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

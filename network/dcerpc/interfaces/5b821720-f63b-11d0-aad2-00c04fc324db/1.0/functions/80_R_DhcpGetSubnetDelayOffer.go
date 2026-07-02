package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpGetSubnetDelayOfferRequest carries the [in] parameters of R_DhcpGetSubnetDelayOffer.
type r_DhcpGetSubnetDelayOfferRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   ndr.DWORD
}

func (*r_DhcpGetSubnetDelayOfferRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpGetSubnetDelayOffer
}

// r_DhcpGetSubnetDelayOfferResponse carries the [out] parameters and return value of R_DhcpGetSubnetDelayOffer.
type r_DhcpGetSubnetDelayOfferResponse struct {
	TimeDelayInMilliseconds uint16
	Status                  ndr.DWORD `ndr:"retval"`
}

// R_DhcpGetSubnetDelayOffer calls R_DhcpGetSubnetDelayOffer (opnum 80) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetSubnetDelayOffer(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD) (TimeDelayInMilliseconds uint16, err error) {
	req := &r_DhcpGetSubnetDelayOfferRequest{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
	}
	var resp r_DhcpGetSubnetDelayOfferResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetSubnetDelayOffer: %w", err)
		return
	}
	TimeDelayInMilliseconds = resp.TimeDelayInMilliseconds
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetSubnetDelayOffer failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

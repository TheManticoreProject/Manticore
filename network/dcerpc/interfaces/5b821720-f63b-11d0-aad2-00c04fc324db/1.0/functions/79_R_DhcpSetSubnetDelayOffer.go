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

// r_DhcpSetSubnetDelayOfferRequest carries the [in] parameters of R_DhcpSetSubnetDelayOffer.
type r_DhcpSetSubnetDelayOfferRequest struct {
	ServerIpAddress         *ndr.WSTR `ndr:"unique"`
	SubnetAddress           ndr.DWORD
	TimeDelayInMilliseconds uint16
}

func (*r_DhcpSetSubnetDelayOfferRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpSetSubnetDelayOffer
}

// r_DhcpSetSubnetDelayOfferResponse carries the [out] parameters and return value of R_DhcpSetSubnetDelayOffer.
type r_DhcpSetSubnetDelayOfferResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetSubnetDelayOffer calls R_DhcpSetSubnetDelayOffer (opnum 79) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetSubnetDelayOffer(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, timeDelayInMilliseconds uint16) (err error) {
	req := &r_DhcpSetSubnetDelayOfferRequest{
		ServerIpAddress:         serverIpAddress,
		SubnetAddress:           subnetAddress,
		TimeDelayInMilliseconds: timeDelayInMilliseconds,
	}
	var resp r_DhcpSetSubnetDelayOfferResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetSubnetDelayOffer: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetSubnetDelayOffer failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

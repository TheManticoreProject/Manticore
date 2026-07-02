package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpV4FailoverGetAddressStatusRequest carries the [in] parameters of R_DhcpV4FailoverGetAddressStatus.
type r_DhcpV4FailoverGetAddressStatusRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   ndr.DWORD
}

func (*r_DhcpV4FailoverGetAddressStatusRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4FailoverGetAddressStatus
}

// r_DhcpV4FailoverGetAddressStatusResponse carries the [out] parameters and return value of R_DhcpV4FailoverGetAddressStatus.
type r_DhcpV4FailoverGetAddressStatusResponse struct {
	PStatus ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4FailoverGetAddressStatus calls R_DhcpV4FailoverGetAddressStatus (opnum 125) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4FailoverGetAddressStatus(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD) (PStatus ndr.DWORD, err error) {
	req := &r_DhcpV4FailoverGetAddressStatusRequest{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
	}
	var resp r_DhcpV4FailoverGetAddressStatusResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4FailoverGetAddressStatus: %w", err)
		return
	}
	PStatus = resp.PStatus
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4FailoverGetAddressStatus failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV6GetStatelessStoreParamsRequest carries the [in] parameters of R_DhcpV6GetStatelessStoreParams.
type r_DhcpV6GetStatelessStoreParamsRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	FServerLevel    ndr.BOOL
	SubnetAddress   msdhcpm.DHCP_IPV6_ADDRESS
}

func (*r_DhcpV6GetStatelessStoreParamsRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV6GetStatelessStoreParams
}

// r_DhcpV6GetStatelessStoreParamsResponse carries the [out] parameters and return value of R_DhcpV6GetStatelessStoreParams.
type r_DhcpV6GetStatelessStoreParamsResponse struct {
	Params msdhcpm.DHCPV6_STATELESS_PARAMS
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV6GetStatelessStoreParams calls R_DhcpV6GetStatelessStoreParams (opnum 117) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV6GetStatelessStoreParams(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, fServerLevel ndr.BOOL, subnetAddress msdhcpm.DHCP_IPV6_ADDRESS) (Params msdhcpm.DHCPV6_STATELESS_PARAMS, err error) {
	req := &r_DhcpV6GetStatelessStoreParamsRequest{
		ServerIpAddress: serverIpAddress,
		FServerLevel:    fServerLevel,
		SubnetAddress:   subnetAddress,
	}
	var resp r_DhcpV6GetStatelessStoreParamsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV6GetStatelessStoreParams: %w", err)
		return
	}
	Params = resp.Params
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV6GetStatelessStoreParams failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

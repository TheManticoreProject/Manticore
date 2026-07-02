package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV6SetStatelessStoreParamsRequest carries the [in] parameters of R_DhcpV6SetStatelessStoreParams.
type r_DhcpV6SetStatelessStoreParamsRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	FServerLevel    ndr.BOOL
	SubnetAddress   msdhcpm.DHCP_IPV6_ADDRESS
	FieldModified   ndr.DWORD
	Params          msdhcpm.DHCPV6_STATELESS_PARAMS
}

func (*r_DhcpV6SetStatelessStoreParamsRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV6SetStatelessStoreParams
}

// r_DhcpV6SetStatelessStoreParamsResponse carries the [out] parameters and return value of R_DhcpV6SetStatelessStoreParams.
type r_DhcpV6SetStatelessStoreParamsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV6SetStatelessStoreParams calls R_DhcpV6SetStatelessStoreParams (opnum 116) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV6SetStatelessStoreParams(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, fServerLevel ndr.BOOL, subnetAddress msdhcpm.DHCP_IPV6_ADDRESS, fieldModified ndr.DWORD, params msdhcpm.DHCPV6_STATELESS_PARAMS) (err error) {
	req := &r_DhcpV6SetStatelessStoreParamsRequest{
		ServerIpAddress: serverIpAddress,
		FServerLevel:    fServerLevel,
		SubnetAddress:   subnetAddress,
		FieldModified:   fieldModified,
		Params:          params,
	}
	var resp r_DhcpV6SetStatelessStoreParamsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV6SetStatelessStoreParams: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV6SetStatelessStoreParams failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

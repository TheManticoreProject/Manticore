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

// r_DhcpV6GetFreeIPAddressRequest carries the [in] parameters of R_DhcpV6GetFreeIPAddress.
type r_DhcpV6GetFreeIPAddressRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ScopeId         msdhcpm.DHCP_IPV6_ADDRESS
	StartIP         msdhcpm.DHCP_IPV6_ADDRESS
	EndIP           msdhcpm.DHCP_IPV6_ADDRESS
	NumFreeAddr     ndr.DWORD
}

func (*r_DhcpV6GetFreeIPAddressRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpV6GetFreeIPAddress }

// r_DhcpV6GetFreeIPAddressResponse carries the [out] parameters and return value of R_DhcpV6GetFreeIPAddress.
type r_DhcpV6GetFreeIPAddressResponse struct {
	IPAddrList *msdhcpm.DHCPV6_IP_ARRAY `ndr:"unique"`
	Status     ndr.DWORD                `ndr:"retval"`
}

// R_DhcpV6GetFreeIPAddress calls R_DhcpV6GetFreeIPAddress (opnum 121) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV6GetFreeIPAddress(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, scopeId msdhcpm.DHCP_IPV6_ADDRESS, startIP msdhcpm.DHCP_IPV6_ADDRESS, endIP msdhcpm.DHCP_IPV6_ADDRESS, numFreeAddr ndr.DWORD) (IPAddrList *msdhcpm.DHCPV6_IP_ARRAY, err error) {
	req := &r_DhcpV6GetFreeIPAddressRequest{
		ServerIpAddress: serverIpAddress,
		ScopeId:         scopeId,
		StartIP:         startIP,
		EndIP:           endIP,
		NumFreeAddr:     numFreeAddr,
	}
	var resp r_DhcpV6GetFreeIPAddressResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV6GetFreeIPAddress: %w", err)
		return
	}
	IPAddrList = resp.IPAddrList
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV6GetFreeIPAddress failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

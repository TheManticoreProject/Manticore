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

// r_DhcpV4GetFreeIPAddressRequest carries the [in] parameters of R_DhcpV4GetFreeIPAddress.
type r_DhcpV4GetFreeIPAddressRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ScopeId         ndr.DWORD
	StartIP         ndr.DWORD
	EndIP           ndr.DWORD
	NumFreeAddr     ndr.DWORD
}

func (*r_DhcpV4GetFreeIPAddressRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpV4GetFreeIPAddress }

// r_DhcpV4GetFreeIPAddressResponse carries the [out] parameters and return value of R_DhcpV4GetFreeIPAddress.
type r_DhcpV4GetFreeIPAddressResponse struct {
	IPAddrList *msdhcpm.DHCP_IP_ARRAY `ndr:"unique"`
	Status     ndr.DWORD              `ndr:"retval"`
}

// R_DhcpV4GetFreeIPAddress calls R_DhcpV4GetFreeIPAddress (opnum 120) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4GetFreeIPAddress(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, scopeId ndr.DWORD, startIP ndr.DWORD, endIP ndr.DWORD, numFreeAddr ndr.DWORD) (IPAddrList *msdhcpm.DHCP_IP_ARRAY, err error) {
	req := &r_DhcpV4GetFreeIPAddressRequest{
		ServerIpAddress: serverIpAddress,
		ScopeId:         scopeId,
		StartIP:         startIP,
		EndIP:           endIP,
		NumFreeAddr:     numFreeAddr,
	}
	var resp r_DhcpV4GetFreeIPAddressResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4GetFreeIPAddress: %w", err)
		return
	}
	IPAddrList = resp.IPAddrList
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4GetFreeIPAddress failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

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

// r_DhcpQueryDnsRegCredentialsRequest carries the [in] parameters of R_DhcpQueryDnsRegCredentials.
type r_DhcpQueryDnsRegCredentialsRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	UnameSize       ndr.DWORD
	DomainSize      ndr.DWORD
}

func (*r_DhcpQueryDnsRegCredentialsRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpQueryDnsRegCredentials
}

// r_DhcpQueryDnsRegCredentialsResponse carries the [out] parameters and return value of R_DhcpQueryDnsRegCredentials.
type r_DhcpQueryDnsRegCredentialsResponse struct {
	Uname  []uint16  `ndr:"ref,size_is=UnameSize"`
	Domain []uint16  `ndr:"ref,size_is=DomainSize"`
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpQueryDnsRegCredentials calls R_DhcpQueryDnsRegCredentials (opnum 42) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpQueryDnsRegCredentials(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, unameSize ndr.DWORD, domainSize ndr.DWORD) (Uname []uint16, Domain []uint16, err error) {
	req := &r_DhcpQueryDnsRegCredentialsRequest{
		ServerIpAddress: serverIpAddress,
		UnameSize:       unameSize,
		DomainSize:      domainSize,
	}
	var resp r_DhcpQueryDnsRegCredentialsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpQueryDnsRegCredentials: %w", err)
		return
	}
	Uname = resp.Uname
	Domain = resp.Domain
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpQueryDnsRegCredentials failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

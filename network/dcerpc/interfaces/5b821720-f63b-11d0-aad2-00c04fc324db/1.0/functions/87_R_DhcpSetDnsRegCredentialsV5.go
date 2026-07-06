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

// r_DhcpSetDnsRegCredentialsV5Request carries the [in] parameters of R_DhcpSetDnsRegCredentialsV5.
type r_DhcpSetDnsRegCredentialsV5Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Uname           *ndr.WSTR `ndr:"unique"`
	Domain          *ndr.WSTR `ndr:"unique"`
	Passwd          *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpSetDnsRegCredentialsV5Request) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpSetDnsRegCredentialsV5
}

// r_DhcpSetDnsRegCredentialsV5Response carries the [out] parameters and return value of R_DhcpSetDnsRegCredentialsV5.
type r_DhcpSetDnsRegCredentialsV5Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetDnsRegCredentialsV5 calls R_DhcpSetDnsRegCredentialsV5 (opnum 87) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetDnsRegCredentialsV5(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, uname *ndr.WSTR, domain *ndr.WSTR, passwd *ndr.WSTR) (err error) {
	req := &r_DhcpSetDnsRegCredentialsV5Request{
		ServerIpAddress: serverIpAddress,
		Uname:           uname,
		Domain:          domain,
		Passwd:          passwd,
	}
	var resp r_DhcpSetDnsRegCredentialsV5Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetDnsRegCredentialsV5: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetDnsRegCredentialsV5 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

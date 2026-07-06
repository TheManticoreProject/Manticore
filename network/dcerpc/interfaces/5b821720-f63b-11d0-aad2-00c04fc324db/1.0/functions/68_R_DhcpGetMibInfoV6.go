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

// r_DhcpGetMibInfoV6Request carries the [in] parameters of R_DhcpGetMibInfoV6.
type r_DhcpGetMibInfoV6Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpGetMibInfoV6Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpGetMibInfoV6 }

// r_DhcpGetMibInfoV6Response carries the [out] parameters and return value of R_DhcpGetMibInfoV6.
type r_DhcpGetMibInfoV6Response struct {
	MibInfo *msdhcpm.DHCP_MIB_INFO_V6 `ndr:"unique"`
	Status  ndr.DWORD                 `ndr:"retval"`
}

// R_DhcpGetMibInfoV6 calls R_DhcpGetMibInfoV6 (opnum 68) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetMibInfoV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR) (MibInfo *msdhcpm.DHCP_MIB_INFO_V6, err error) {
	req := &r_DhcpGetMibInfoV6Request{
		ServerIpAddress: serverIpAddress,
	}
	var resp r_DhcpGetMibInfoV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetMibInfoV6: %w", err)
		return
	}
	MibInfo = resp.MibInfo
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetMibInfoV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

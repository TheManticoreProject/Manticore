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

// r_DhcpServerQueryAttributesRequest carries the [in] parameters of R_DhcpServerQueryAttributes.
type r_DhcpServerQueryAttributesRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	DwReserved      ndr.DWORD
	DwAttribCount   ndr.DWORD
	PDhcpAttribs    []ndr.DWORD `ndr:"ref,size_is=DwAttribCount"`
}

func (*r_DhcpServerQueryAttributesRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpServerQueryAttributes
}

// r_DhcpServerQueryAttributesResponse carries the [out] parameters and return value of R_DhcpServerQueryAttributes.
type r_DhcpServerQueryAttributesResponse struct {
	PDhcpAttribArr *msdhcpm.DHCP_ATTRIB_ARRAY `ndr:"unique"`
	Status         ndr.DWORD                  `ndr:"retval"`
}

// R_DhcpServerQueryAttributes calls R_DhcpServerQueryAttributes (opnum 35) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpServerQueryAttributes(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, dwReserved ndr.DWORD, dwAttribCount ndr.DWORD, pDhcpAttribs []ndr.DWORD) (PDhcpAttribArr *msdhcpm.DHCP_ATTRIB_ARRAY, err error) {
	req := &r_DhcpServerQueryAttributesRequest{
		ServerIpAddress: serverIpAddress,
		DwReserved:      dwReserved,
		DwAttribCount:   dwAttribCount,
		PDhcpAttribs:    pDhcpAttribs,
	}
	var resp r_DhcpServerQueryAttributesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpServerQueryAttributes: %w", err)
		return
	}
	PDhcpAttribArr = resp.PDhcpAttribArr
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpServerQueryAttributes failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

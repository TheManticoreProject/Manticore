package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpServerQueryAttributeRequest carries the [in] parameters of R_DhcpServerQueryAttribute.
type r_DhcpServerQueryAttributeRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	DwReserved      ndr.DWORD
	DhcpAttribId    ndr.DWORD
}

func (*r_DhcpServerQueryAttributeRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpServerQueryAttribute
}

// r_DhcpServerQueryAttributeResponse carries the [out] parameters and return value of R_DhcpServerQueryAttribute.
type r_DhcpServerQueryAttributeResponse struct {
	PDhcpAttrib *msdhcpm.DHCP_ATTRIB `ndr:"unique"`
	Status      ndr.DWORD            `ndr:"retval"`
}

// R_DhcpServerQueryAttribute calls R_DhcpServerQueryAttribute (opnum 34) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpServerQueryAttribute(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, dwReserved ndr.DWORD, dhcpAttribId ndr.DWORD) (PDhcpAttrib *msdhcpm.DHCP_ATTRIB, err error) {
	req := &r_DhcpServerQueryAttributeRequest{
		ServerIpAddress: serverIpAddress,
		DwReserved:      dwReserved,
		DhcpAttribId:    dhcpAttribId,
	}
	var resp r_DhcpServerQueryAttributeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpServerQueryAttribute: %w", err)
		return
	}
	PDhcpAttrib = resp.PDhcpAttrib
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpServerQueryAttribute failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

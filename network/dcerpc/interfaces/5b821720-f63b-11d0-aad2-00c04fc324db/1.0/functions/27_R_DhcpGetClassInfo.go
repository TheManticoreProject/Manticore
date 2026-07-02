package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpGetClassInfoRequest carries the [in] parameters of R_DhcpGetClassInfo.
type r_DhcpGetClassInfoRequest struct {
	ServerIpAddress    *ndr.WSTR `ndr:"unique"`
	ReservedMustBeZero ndr.DWORD
	PartialClassInfo   msdhcpm.DHCP_CLASS_INFO
}

func (*r_DhcpGetClassInfoRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpGetClassInfo }

// r_DhcpGetClassInfoResponse carries the [out] parameters and return value of R_DhcpGetClassInfo.
type r_DhcpGetClassInfoResponse struct {
	FilledClassInfo *msdhcpm.DHCP_CLASS_INFO `ndr:"unique"`
	Status          ndr.DWORD                `ndr:"retval"`
}

// R_DhcpGetClassInfo calls R_DhcpGetClassInfo (opnum 27) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetClassInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, reservedMustBeZero ndr.DWORD, partialClassInfo msdhcpm.DHCP_CLASS_INFO) (FilledClassInfo *msdhcpm.DHCP_CLASS_INFO, err error) {
	req := &r_DhcpGetClassInfoRequest{
		ServerIpAddress:    serverIpAddress,
		ReservedMustBeZero: reservedMustBeZero,
		PartialClassInfo:   partialClassInfo,
	}
	var resp r_DhcpGetClassInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetClassInfo: %w", err)
		return
	}
	FilledClassInfo = resp.FilledClassInfo
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetClassInfo failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

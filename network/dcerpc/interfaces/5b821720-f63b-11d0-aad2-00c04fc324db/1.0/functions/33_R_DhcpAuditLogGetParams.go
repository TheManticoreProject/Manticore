package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpAuditLogGetParamsRequest carries the [in] parameters of R_DhcpAuditLogGetParams.
type r_DhcpAuditLogGetParamsRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
}

func (*r_DhcpAuditLogGetParamsRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpAuditLogGetParams }

// r_DhcpAuditLogGetParamsResponse carries the [out] parameters and return value of R_DhcpAuditLogGetParams.
type r_DhcpAuditLogGetParamsResponse struct {
	AuditLogDir       *ndr.WSTR `ndr:"unique"`
	DiskCheckInterval ndr.DWORD
	MaxLogFilesSize   ndr.DWORD
	MinSpaceOnDisk    ndr.DWORD
	Status            ndr.DWORD `ndr:"retval"`
}

// R_DhcpAuditLogGetParams calls R_DhcpAuditLogGetParams (opnum 33) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpAuditLogGetParams(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD) (AuditLogDir *ndr.WSTR, DiskCheckInterval ndr.DWORD, MaxLogFilesSize ndr.DWORD, MinSpaceOnDisk ndr.DWORD, err error) {
	req := &r_DhcpAuditLogGetParamsRequest{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
	}
	var resp r_DhcpAuditLogGetParamsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpAuditLogGetParams: %w", err)
		return
	}
	AuditLogDir = resp.AuditLogDir
	DiskCheckInterval = resp.DiskCheckInterval
	MaxLogFilesSize = resp.MaxLogFilesSize
	MinSpaceOnDisk = resp.MinSpaceOnDisk
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpAuditLogGetParams failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

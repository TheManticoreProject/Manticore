package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpAuditLogSetParamsRequest carries the [in] parameters of R_DhcpAuditLogSetParams.
type r_DhcpAuditLogSetParamsRequest struct {
	ServerIpAddress   *ndr.WSTR `ndr:"unique"`
	Flags             ndr.DWORD
	AuditLogDir       ndr.WSTR
	DiskCheckInterval ndr.DWORD
	MaxLogFilesSize   ndr.DWORD
	MinSpaceOnDisk    ndr.DWORD
}

func (*r_DhcpAuditLogSetParamsRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpAuditLogSetParams }

// r_DhcpAuditLogSetParamsResponse carries the [out] parameters and return value of R_DhcpAuditLogSetParams.
type r_DhcpAuditLogSetParamsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpAuditLogSetParams calls R_DhcpAuditLogSetParams (opnum 32) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpAuditLogSetParams(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, auditLogDir ndr.WSTR, diskCheckInterval ndr.DWORD, maxLogFilesSize ndr.DWORD, minSpaceOnDisk ndr.DWORD) (err error) {
	req := &r_DhcpAuditLogSetParamsRequest{
		ServerIpAddress:   serverIpAddress,
		Flags:             flags,
		AuditLogDir:       auditLogDir,
		DiskCheckInterval: diskCheckInterval,
		MaxLogFilesSize:   maxLogFilesSize,
		MinSpaceOnDisk:    minSpaceOnDisk,
	}
	var resp r_DhcpAuditLogSetParamsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpAuditLogSetParams: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpAuditLogSetParams failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpScanDatabaseRequest carries the [in] parameters of R_DhcpScanDatabase.
type r_DhcpScanDatabaseRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   ndr.DWORD
	FixFlag         ndr.DWORD
}

func (*r_DhcpScanDatabaseRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpScanDatabase }

// r_DhcpScanDatabaseResponse carries the [out] parameters and return value of R_DhcpScanDatabase.
type r_DhcpScanDatabaseResponse struct {
	ScanList *msdhcpm.DHCP_SCAN_LIST `ndr:"unique"`
	Status   ndr.DWORD               `ndr:"retval"`
}

// R_DhcpScanDatabase calls R_DhcpScanDatabase (opnum 27) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpScanDatabase(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, fixFlag ndr.DWORD) (ScanList *msdhcpm.DHCP_SCAN_LIST, err error) {
	req := &r_DhcpScanDatabaseRequest{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
		FixFlag:         fixFlag,
	}
	var resp r_DhcpScanDatabaseResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpScanDatabase: %w", err)
		return
	}
	ScanList = resp.ScanList
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpScanDatabase failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpScanMDatabaseRequest carries the [in] parameters of R_DhcpScanMDatabase.
type r_DhcpScanMDatabaseRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	MScopeName      ndr.WSTR
	FixFlag         ndr.DWORD
}

func (*r_DhcpScanMDatabaseRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpScanMDatabase }

// r_DhcpScanMDatabaseResponse carries the [out] parameters and return value of R_DhcpScanMDatabase.
type r_DhcpScanMDatabaseResponse struct {
	ScanList *msdhcpm.DHCP_SCAN_LIST `ndr:"unique"`
	Status   ndr.DWORD               `ndr:"retval"`
}

// R_DhcpScanMDatabase calls R_DhcpScanMDatabase (opnum 8) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpScanMDatabase(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, mScopeName ndr.WSTR, fixFlag ndr.DWORD) (ScanList *msdhcpm.DHCP_SCAN_LIST, err error) {
	req := &r_DhcpScanMDatabaseRequest{
		ServerIpAddress: serverIpAddress,
		MScopeName:      mScopeName,
		FixFlag:         fixFlag,
	}
	var resp r_DhcpScanMDatabaseResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpScanMDatabase: %w", err)
		return
	}
	ScanList = resp.ScanList
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpScanMDatabase failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

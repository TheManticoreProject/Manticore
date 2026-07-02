package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpV4FailoverTriggerAddrAllocationRequest carries the [in] parameters of R_DhcpV4FailoverTriggerAddrAllocation.
type r_DhcpV4FailoverTriggerAddrAllocationRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	FailRelName     *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpV4FailoverTriggerAddrAllocationRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4FailoverTriggerAddrAllocation
}

// r_DhcpV4FailoverTriggerAddrAllocationResponse carries the [out] parameters and return value of R_DhcpV4FailoverTriggerAddrAllocation.
type r_DhcpV4FailoverTriggerAddrAllocationResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4FailoverTriggerAddrAllocation calls R_DhcpV4FailoverTriggerAddrAllocation (opnum 100) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4FailoverTriggerAddrAllocation(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, failRelName *ndr.WSTR) (err error) {
	req := &r_DhcpV4FailoverTriggerAddrAllocationRequest{
		ServerIpAddress: serverIpAddress,
		FailRelName:     failRelName,
	}
	var resp r_DhcpV4FailoverTriggerAddrAllocationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4FailoverTriggerAddrAllocation: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4FailoverTriggerAddrAllocation failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}

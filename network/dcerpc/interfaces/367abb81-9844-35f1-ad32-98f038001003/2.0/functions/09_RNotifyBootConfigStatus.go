package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rNotifyBootConfigStatusRequest carries the [in] parameters of RNotifyBootConfigStatus.
type rNotifyBootConfigStatusRequest struct {
	LpMachineName  *ndr.WSTR `ndr:"unique"`
	BootAcceptable ndr.DWORD
}

func (*rNotifyBootConfigStatusRequest) Opnum() uint16 { return svcctl.OpnumRNotifyBootConfigStatus }

// rNotifyBootConfigStatusResponse carries the [out] parameters and return value of RNotifyBootConfigStatus.
type rNotifyBootConfigStatusResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RNotifyBootConfigStatus calls RNotifyBootConfigStatus (opnum 9) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RNotifyBootConfigStatus(rpc ndr.Invoker, lpMachineName *ndr.WSTR, bootAcceptable ndr.DWORD) (err error) {
	req := &rNotifyBootConfigStatusRequest{
		LpMachineName:  lpMachineName,
		BootAcceptable: bootAcceptable,
	}
	var resp rNotifyBootConfigStatusResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RNotifyBootConfigStatus: %w", err)
		return
	}
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RNotifyBootConfigStatus failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}

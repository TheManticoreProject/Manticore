package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcEnumMonitorsRequest carries the [in] parameters of RpcEnumMonitors.
type rpcEnumMonitorsRequest struct {
	PName    *ndr.WSTR `ndr:"unique"`
	Level    ndr.DWORD
	PMonitor []uint8 `ndr:"unique,size_is=CbBuf"`
	CbBuf    ndr.DWORD
}

func (*rpcEnumMonitorsRequest) Opnum() uint16 { return winspool.OpnumRpcEnumMonitors }

// rpcEnumMonitorsResponse carries the [out] parameters and return value of RpcEnumMonitors.
type rpcEnumMonitorsResponse struct {
	PMonitor   []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded  ndr.DWORD
	PcReturned ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// RpcEnumMonitors calls RpcEnumMonitors (opnum 36) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcEnumMonitors(rpc ndr.Invoker, pName *ndr.WSTR, level ndr.DWORD, pMonitor []uint8, cbBuf ndr.DWORD) (PMonitor []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcEnumMonitorsRequest{
		PName:    pName,
		Level:    level,
		PMonitor: pMonitor,
		CbBuf:    cbBuf,
	}
	var resp rpcEnumMonitorsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcEnumMonitors: %w", err)
		return
	}
	PMonitor = resp.PMonitor
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcEnumMonitors failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}

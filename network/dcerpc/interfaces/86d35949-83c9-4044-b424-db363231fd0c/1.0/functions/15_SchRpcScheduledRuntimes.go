package functions

import (
	"fmt"

	schrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/86d35949-83c9-4044-b424-db363231fd0c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// schRpcScheduledRuntimesRequest carries the [in] parameters of SchRpcScheduledRuntimes.
type schRpcScheduledRuntimesRequest struct {
	Path       ndr.WSTR
	Start      *msdtyp.SYSTEMTIME `ndr:"unique"`
	End        *msdtyp.SYSTEMTIME `ndr:"unique"`
	Flags      ndr.DWORD
	CRequested ndr.DWORD
}

func (*schRpcScheduledRuntimesRequest) Opnum() uint16 {
	return schrpc.OpnumSchRpcScheduledRuntimes
}

// schRpcScheduledRuntimesResponse carries the [out] parameters and return value of SchRpcScheduledRuntimes.
type schRpcScheduledRuntimesResponse struct {
	PcRuntimes ndr.DWORD
	PRuntimes  []msdtyp.SYSTEMTIME `ndr:"unique,size_is=PcRuntimes"`
	Status     ndr.DWORD           `ndr:"retval"`
}

// SchRpcScheduledRuntimes calls SchRpcScheduledRuntimes (opnum 15) ([MS-TSCH] section 3.2.5.4.16).
func SchRpcScheduledRuntimes(rpc ndr.Invoker, path ndr.WSTR, start *msdtyp.SYSTEMTIME, end *msdtyp.SYSTEMTIME, flags ndr.DWORD, cRequested ndr.DWORD) (PcRuntimes ndr.DWORD, PRuntimes []msdtyp.SYSTEMTIME, err error) {
	req := &schRpcScheduledRuntimesRequest{
		Path:       path,
		Start:      start,
		End:        end,
		Flags:      flags,
		CRequested: cRequested,
	}
	var resp schRpcScheduledRuntimesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcScheduledRuntimes: %w", err)
		return
	}
	PcRuntimes = resp.PcRuntimes
	PRuntimes = resp.PRuntimes
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcScheduledRuntimes failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}

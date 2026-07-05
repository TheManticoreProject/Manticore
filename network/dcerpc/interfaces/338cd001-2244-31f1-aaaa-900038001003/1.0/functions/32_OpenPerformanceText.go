package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// openPerformanceTextRequest carries the [in] parameters of OpenPerformanceText.
type openPerformanceTextRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	SamDesired ndr.DWORD
}

func (*openPerformanceTextRequest) Opnum() uint16 { return winreg.OpnumOpenPerformanceText }

// openPerformanceTextResponse carries the [out] parameters and return value of OpenPerformanceText.
type openPerformanceTextResponse struct {
	PhKey  msrrp.PRPC_HKEY
	Status ndr.DWORD `ndr:"retval"`
}

// OpenPerformanceText calls OpenPerformanceText (opnum 32) ([MS-RRP] — verify the parameter
// modeling and status handling).
func OpenPerformanceText(rpc ndr.Invoker, serverName *ndr.WSTR, samDesired ndr.DWORD) (PhKey msrrp.PRPC_HKEY, err error) {
	req := &openPerformanceTextRequest{
		ServerName: serverName,
		SamDesired: samDesired,
	}
	var resp openPerformanceTextResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("OpenPerformanceText: %w", err)
		return
	}
	PhKey = resp.PhKey
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("OpenPerformanceText failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}

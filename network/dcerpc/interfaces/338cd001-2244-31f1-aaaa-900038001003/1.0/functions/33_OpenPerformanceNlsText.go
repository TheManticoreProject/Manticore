package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// openPerformanceNlsTextRequest carries the [in] parameters of OpenPerformanceNlsText.
type openPerformanceNlsTextRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	SamDesired ndr.DWORD
}

func (*openPerformanceNlsTextRequest) Opnum() uint16 { return winreg.OpnumOpenPerformanceNlsText }

// openPerformanceNlsTextResponse carries the [out] parameters and return value of OpenPerformanceNlsText.
type openPerformanceNlsTextResponse struct {
	PhKey  msrrp.PRPC_HKEY
	Status ndr.DWORD `ndr:"retval"`
}

// OpenPerformanceNlsText calls OpenPerformanceNlsText (opnum 33) ([MS-RRP] — verify the parameter
// modeling and status handling).
func OpenPerformanceNlsText(rpc ndr.Invoker, serverName *ndr.WSTR, samDesired ndr.DWORD) (PhKey msrrp.PRPC_HKEY, err error) {
	req := &openPerformanceNlsTextRequest{
		ServerName: serverName,
		SamDesired: samDesired,
	}
	var resp openPerformanceNlsTextResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("OpenPerformanceNlsText: %w", err)
		return
	}
	PhKey = resp.PhKey
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("OpenPerformanceNlsText failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}

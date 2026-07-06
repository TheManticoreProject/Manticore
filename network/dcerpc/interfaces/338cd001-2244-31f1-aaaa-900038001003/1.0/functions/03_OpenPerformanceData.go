package functions

// IDL source: [MS-RRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrp/47f3edf6-4c2d-45d8-ab5b-2dc077738903
// A fetched copy is kept at ms-rrp.idl in the interface directory.

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// openPerformanceDataRequest carries the [in] parameters of OpenPerformanceData.
type openPerformanceDataRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	SamDesired ndr.DWORD
}

func (*openPerformanceDataRequest) Opnum() uint16 { return winreg.OpnumOpenPerformanceData }

// openPerformanceDataResponse carries the [out] parameters and return value of OpenPerformanceData.
type openPerformanceDataResponse struct {
	PhKey  msrrp.PRPC_HKEY
	Status ndr.DWORD `ndr:"retval"`
}

// OpenPerformanceData calls OpenPerformanceData (opnum 3) ([MS-RRP] — verify the parameter
// modeling and status handling).
func OpenPerformanceData(rpc ndr.Invoker, serverName *ndr.WSTR, samDesired ndr.DWORD) (PhKey msrrp.PRPC_HKEY, err error) {
	req := &openPerformanceDataRequest{
		ServerName: serverName,
		SamDesired: samDesired,
	}
	var resp openPerformanceDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("OpenPerformanceData: %w", err)
		return
	}
	PhKey = resp.PhKey
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("OpenPerformanceData failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}

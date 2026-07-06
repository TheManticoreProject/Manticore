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

// openLocalMachineRequest carries the [in] parameters of OpenLocalMachine.
type openLocalMachineRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	SamDesired ndr.DWORD
}

func (*openLocalMachineRequest) Opnum() uint16 { return winreg.OpnumOpenLocalMachine }

// openLocalMachineResponse carries the [out] parameters and return value of OpenLocalMachine.
type openLocalMachineResponse struct {
	PhKey  msrrp.PRPC_HKEY
	Status ndr.DWORD `ndr:"retval"`
}

// OpenLocalMachine calls OpenLocalMachine (opnum 2) ([MS-RRP] — verify the parameter
// modeling and status handling).
func OpenLocalMachine(rpc ndr.Invoker, serverName *ndr.WSTR, samDesired ndr.DWORD) (PhKey msrrp.PRPC_HKEY, err error) {
	req := &openLocalMachineRequest{
		ServerName: serverName,
		SamDesired: samDesired,
	}
	var resp openLocalMachineResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("OpenLocalMachine: %w", err)
		return
	}
	PhKey = resp.PhKey
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("OpenLocalMachine failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}

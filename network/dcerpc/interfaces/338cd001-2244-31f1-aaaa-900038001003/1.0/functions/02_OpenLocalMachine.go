package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// openLocalMachineRequest carries the [in] parameters of OpenLocalMachine.
type openLocalMachineRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	SamDesired ndr.DWORD
}

func (*openLocalMachineRequest) Opnum() uint16 { return winreg.OpnumOpenLocalMachine }

// openLocalMachineResponse carries the [out] parameters and return value of OpenLocalMachine.
type openLocalMachineResponse struct {
	PhKey  structures.PRPC_HKEY
	Status ndr.DWORD `ndr:"retval"`
}

// OpenLocalMachine calls OpenLocalMachine (opnum 2) ([MS-RRP] — verify the parameter
// modeling and status handling).
func OpenLocalMachine(rpc ndr.Invoker, serverName *ndr.WSTR, samDesired ndr.DWORD) (PhKey structures.PRPC_HKEY, err error) {
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

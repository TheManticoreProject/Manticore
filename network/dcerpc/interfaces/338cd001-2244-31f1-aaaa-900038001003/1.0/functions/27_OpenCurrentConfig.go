package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// openCurrentConfigRequest carries the [in] parameters of OpenCurrentConfig.
type openCurrentConfigRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	SamDesired ndr.DWORD
}

func (*openCurrentConfigRequest) Opnum() uint16 { return winreg.OpnumOpenCurrentConfig }

// openCurrentConfigResponse carries the [out] parameters and return value of OpenCurrentConfig.
type openCurrentConfigResponse struct {
	PhKey  structures.PRPC_HKEY
	Status ndr.DWORD `ndr:"retval"`
}

// OpenCurrentConfig calls OpenCurrentConfig (opnum 27) ([MS-RRP] — verify the parameter
// modeling and status handling).
func OpenCurrentConfig(rpc ndr.Invoker, serverName *ndr.WSTR, samDesired ndr.DWORD) (PhKey structures.PRPC_HKEY, err error) {
	req := &openCurrentConfigRequest{
		ServerName: serverName,
		SamDesired: samDesired,
	}
	var resp openCurrentConfigResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("OpenCurrentConfig: %w", err)
		return
	}
	PhKey = resp.PhKey
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("OpenCurrentConfig failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}

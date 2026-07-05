package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// openCurrentUserRequest carries the [in] parameters of OpenCurrentUser.
type openCurrentUserRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	SamDesired ndr.DWORD
}

func (*openCurrentUserRequest) Opnum() uint16 { return winreg.OpnumOpenCurrentUser }

// openCurrentUserResponse carries the [out] parameters and return value of OpenCurrentUser.
type openCurrentUserResponse struct {
	PhKey  msrrp.PRPC_HKEY
	Status ndr.DWORD `ndr:"retval"`
}

// OpenCurrentUser calls OpenCurrentUser (opnum 1) ([MS-RRP] — verify the parameter
// modeling and status handling).
func OpenCurrentUser(rpc ndr.Invoker, serverName *ndr.WSTR, samDesired ndr.DWORD) (PhKey msrrp.PRPC_HKEY, err error) {
	req := &openCurrentUserRequest{
		ServerName: serverName,
		SamDesired: samDesired,
	}
	var resp openCurrentUserResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("OpenCurrentUser: %w", err)
		return
	}
	PhKey = resp.PhKey
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("OpenCurrentUser failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}

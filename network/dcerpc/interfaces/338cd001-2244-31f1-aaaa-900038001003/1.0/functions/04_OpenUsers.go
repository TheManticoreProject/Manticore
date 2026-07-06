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

// openUsersRequest carries the [in] parameters of OpenUsers.
type openUsersRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	SamDesired ndr.DWORD
}

func (*openUsersRequest) Opnum() uint16 { return winreg.OpnumOpenUsers }

// openUsersResponse carries the [out] parameters and return value of OpenUsers.
type openUsersResponse struct {
	PhKey  msrrp.PRPC_HKEY
	Status ndr.DWORD `ndr:"retval"`
}

// OpenUsers calls OpenUsers (opnum 4) ([MS-RRP] — verify the parameter
// modeling and status handling).
func OpenUsers(rpc ndr.Invoker, serverName *ndr.WSTR, samDesired ndr.DWORD) (PhKey msrrp.PRPC_HKEY, err error) {
	req := &openUsersRequest{
		ServerName: serverName,
		SamDesired: samDesired,
	}
	var resp openUsersResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("OpenUsers: %w", err)
		return
	}
	PhKey = resp.PhKey
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("OpenUsers failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	msgsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/17fdd703-1827-4e34-79d4-24a55c53bb37/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrMessageNameDelRequest carries the [in] parameters of NetrMessageNameDel.
type netrMessageNameDelRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	MsgName    ndr.WSTR
}

func (*netrMessageNameDelRequest) Opnum() uint16 { return msgsvc.OpnumNetrMessageNameDel }

// netrMessageNameDelResponse carries the [out] parameters and return value of NetrMessageNameDel.
type netrMessageNameDelResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrMessageNameDel calls NetrMessageNameDel (opnum 3) ([MS-MSRP] — verify the parameter
// modeling and status handling).
func NetrMessageNameDel(rpc ndr.Invoker, serverName *ndr.WSTR, msgName ndr.WSTR) (err error) {
	req := &netrMessageNameDelRequest{
		ServerName: serverName,
		MsgName:    msgName,
	}
	var resp netrMessageNameDelResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrMessageNameDel: %w", err)
		return
	}
	if uint32(resp.Status) != msgsvc.StatusSuccess {
		err = fmt.Errorf("NetrMessageNameDel failed: %s", msgsvc.StatusString(uint32(resp.Status)))
	}
	return
}

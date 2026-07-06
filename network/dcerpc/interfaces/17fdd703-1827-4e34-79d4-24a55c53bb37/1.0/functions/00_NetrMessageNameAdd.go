package functions

// IDL source: [MS-MSRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-msrp/181965ff-fab4-4ad4-a8d7-16b444cc4e66
// A fetched copy is kept at ms-msrp.idl in the interface directory.

import (
	"fmt"

	msgsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/17fdd703-1827-4e34-79d4-24a55c53bb37/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrMessageNameAddRequest carries the [in] parameters of NetrMessageNameAdd.
type netrMessageNameAddRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	MsgName    ndr.WSTR
}

func (*netrMessageNameAddRequest) Opnum() uint16 { return msgsvc.OpnumNetrMessageNameAdd }

// netrMessageNameAddResponse carries the [out] parameters and return value of NetrMessageNameAdd.
type netrMessageNameAddResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrMessageNameAdd calls NetrMessageNameAdd (opnum 0) ([MS-MSRP] — verify the parameter
// modeling and status handling).
func NetrMessageNameAdd(rpc ndr.Invoker, serverName *ndr.WSTR, msgName ndr.WSTR) (err error) {
	req := &netrMessageNameAddRequest{
		ServerName: serverName,
		MsgName:    msgName,
	}
	var resp netrMessageNameAddResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrMessageNameAdd: %w", err)
		return
	}
	if uint32(resp.Status) != msgsvc.StatusSuccess {
		err = fmt.Errorf("NetrMessageNameAdd failed: %s", msgsvc.StatusString(uint32(resp.Status)))
	}
	return
}

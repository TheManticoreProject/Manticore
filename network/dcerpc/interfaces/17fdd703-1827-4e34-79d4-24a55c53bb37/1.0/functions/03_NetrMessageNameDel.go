package functions

// IDL source: [MS-MSRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-msrp/181965ff-fab4-4ad4-a8d7-16b444cc4e66
// A fetched copy is kept at ms-msrp.idl in the interface directory.

import (
	"fmt"

	msgsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/17fdd703-1827-4e34-79d4-24a55c53bb37/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
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
	if win32.WIN32_ERROR(resp.Status) != win32.NERR_Success {
		err = fmt.Errorf("NetrMessageNameDel failed: %s", win32.WIN32_ERROR(resp.Status).String())
	}
	return
}

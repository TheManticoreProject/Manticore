package functions

// IDL source: [MS-SRVS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/77aacc74-f8f9-4b46-b2d8-bfe04a7d9c44
// A fetched copy is kept at ms-srvs.idl in the interface directory.

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrShareDelRequest is the [in] parameter set of NetrShareDel: the optional server name,
// the [in,string] (ref) share name, and a reserved DWORD (MUST be zero).
type netrShareDelRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	NetName    ndr.WSTR
	Reserved   ndr.DWORD
}

func (*netrShareDelRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrShareDel
}

// NetrShareDel calls NetrShareDel (opnum 18), deleting a share from the server
// ([MS-SRVS] 3.1.4.12).
func NetrShareDel(rpc ndr.Invoker, serverName string, netName string, reserved ndr.DWORD) error {
	req := &netrShareDelRequest{
		ServerName: optWStr(serverName),
		NetName:    ndr.WSTR(netName),
		Reserved:   reserved,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrShareDel: %w", err)
	}
	if uint32(resp.Status) != srvsvc.NERR_Success && uint32(resp.Status) != srvsvc.ERROR_MORE_DATA {
		return fmt.Errorf("NetrShareDel failed: %s", srvsvc.StatusString(uint32(resp.Status)))
	}
	return nil
}

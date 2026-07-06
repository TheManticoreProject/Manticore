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

// netrFileCloseRequest is the [in] parameter set of NetrFileClose: the [unique] server
// name and the identifier of the open file/resource to force closed.
type netrFileCloseRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	FileId     ndr.DWORD
}

func (*netrFileCloseRequest) Opnum() uint16 { return srvsvc.OpnumNetrFileClose }

// NetrFileClose calls NetrFileClose (opnum 11), forcing the open file/resource
// identified by FileId closed ([MS-SRVS] 3.1.4.4).
func NetrFileClose(rpc ndr.Invoker, serverName string, fileId uint32) error {
	req := &netrFileCloseRequest{
		ServerName: optWStr(serverName),
		FileId:     ndr.DWORD(fileId),
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrFileClose: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return fmt.Errorf("NetrFileClose failed: %s", srvsvc.StatusString(status))
	}
	return nil
}

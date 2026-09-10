package functions

// IDL source: [MS-SRVS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/77aacc74-f8f9-4b46-b2d8-bfe04a7d9c44
// A fetched copy is kept at ms-srvs.idl in the interface directory.

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// netrDfsGetVersionRequest is the [in] parameter set of NetrDfsGetVersion: the [unique]
// server name.
type netrDfsGetVersionRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
}

func (*netrDfsGetVersionRequest) Opnum() uint16 { return srvsvc.OpnumNetrDfsGetVersion }

// netrDfsGetVersionResponse is the reply: the [out] DFS version and the NET_API_STATUS
// return value.
type netrDfsGetVersionResponse struct {
	Version ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// NetrDfsGetVersion calls NetrDfsGetVersion (opnum 43), checking whether the server is a
// DFS server and returning its DFS version ([MS-SRVS] 3.1.4.44).
func NetrDfsGetVersion(rpc ndr.Invoker, serverName string) (uint32, error) {
	req := &netrDfsGetVersionRequest{
		ServerName: optWStr(serverName),
	}
	var resp netrDfsGetVersionResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return 0, fmt.Errorf("NetrDfsGetVersion: %w", err)
	}
	status := win32.WIN32_ERROR(resp.Status)
	if status != win32.NERR_Success && status != win32.ERROR_MORE_DATA {
		return uint32(resp.Version), fmt.Errorf("NetrDfsGetVersion failed: %s", status)
	}
	return uint32(resp.Version), nil
}

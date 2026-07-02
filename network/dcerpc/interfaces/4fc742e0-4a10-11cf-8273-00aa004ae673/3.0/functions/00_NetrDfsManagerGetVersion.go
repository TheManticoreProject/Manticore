package functions

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrDfsManagerGetVersionRequest carries the [in] parameters of NetrDfsManagerGetVersion.
type netrDfsManagerGetVersionRequest struct {
}

func (*netrDfsManagerGetVersionRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsManagerGetVersion }

// netrDfsManagerGetVersionResponse carries the return value of NetrDfsManagerGetVersion.
// Unlike the other netdfs methods, the DWORD return value is the DFS server's supported
// version number ([MS-DFSNM] 3.1.4.1.1), NOT a NET_API_STATUS error code, so every value
// (including 0) is a valid result and there is no status to check.
type netrDfsManagerGetVersionResponse struct {
	Version ndr.DWORD `ndr:"retval"`
}

// NetrDfsManagerGetVersion calls NetrDfsManagerGetVersion (opnum 0) ([MS-DFSNM] 3.1.4.1.1)
// and returns the DFS version supported by the server. A client uses this value to decide
// which DFS methods the server understands; it is not an error code.
func NetrDfsManagerGetVersion(rpc ndr.Invoker) (version uint32, err error) {
	req := &netrDfsManagerGetVersionRequest{}
	var resp netrDfsManagerGetVersionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsManagerGetVersion: %w", err)
		return
	}
	version = uint32(resp.Version)
	return
}

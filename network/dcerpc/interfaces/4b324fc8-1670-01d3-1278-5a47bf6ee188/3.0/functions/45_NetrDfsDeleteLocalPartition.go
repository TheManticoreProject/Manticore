package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// netrDfsDeleteLocalPartitionRequest is the [in] parameter set of
// NetrDfsDeleteLocalPartition: the [unique] server name, the inline partition GUID (a
// single [in] pointer in the IDL), and the (ref) prefix.
type netrDfsDeleteLocalPartitionRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Uid        guid.GUID
	Prefix     ndr.WSTR
}

func (*netrDfsDeleteLocalPartitionRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrDfsDeleteLocalPartition
}

// NetrDfsDeleteLocalPartition calls NetrDfsDeleteLocalPartition (opnum 45), deleting a
// local DFS partition ([MS-SRVS] 3.1.4.46).
func NetrDfsDeleteLocalPartition(rpc ndr.Invoker, serverName string, uid guid.GUID, prefix string) error {
	req := &netrDfsDeleteLocalPartitionRequest{
		ServerName: optWStr(serverName),
		Uid:        uid,
		Prefix:     ndr.WSTR(prefix),
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrDfsDeleteLocalPartition: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return fmt.Errorf("NetrDfsDeleteLocalPartition failed: %s", srvsvc.StatusString(status))
	}
	return nil
}

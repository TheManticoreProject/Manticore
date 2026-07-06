package functions

// IDL source: [MS-SRVS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/77aacc74-f8f9-4b46-b2d8-bfe04a7d9c44
// A fetched copy is kept at ms-srvs.idl in the interface directory.

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// netrDfsDeleteLocalPartitionRequest is the [in] parameter set of
// NetrDfsDeleteLocalPartition: the [unique] server name, the inline partition GUID (a
// single [in] pointer in the IDL), and the (ref) prefix.
type netrDfsDeleteLocalPartitionRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Uid        msdtyp.GUID
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
		Uid:        msdtyp.NewGUID(uid),
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

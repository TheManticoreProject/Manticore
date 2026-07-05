package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	mssrvs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// netrDfsCreateLocalPartitionRequest is the [in] parameter set of
// NetrDfsCreateLocalPartition: the [unique] server name, the (ref) share name, the inline
// entry GUID (a single [in] pointer in the IDL), the (ref) entry prefix, the (ref) short
// name, the inline NET_DFS_ENTRY_ID_CONTAINER relation info (a single [in] ref pointer),
// and the Force flag.
type netrDfsCreateLocalPartitionRequest struct {
	ServerName   *ndr.WSTR `ndr:"unique"`
	ShareName    ndr.WSTR
	EntryUid     dtyp.GUID
	EntryPrefix  ndr.WSTR
	ShortName    ndr.WSTR
	RelationInfo mssrvs.NET_DFS_ENTRY_ID_CONTAINER
	Force        int32
}

func (*netrDfsCreateLocalPartitionRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrDfsCreateLocalPartition
}

// NetrDfsCreateLocalPartition calls NetrDfsCreateLocalPartition (opnum 44), creating a
// local DFS partition ([MS-SRVS] 3.1.4.45).
func NetrDfsCreateLocalPartition(rpc ndr.Invoker, serverName, shareName string, entryUid guid.GUID, entryPrefix, shortName string, relationInfo mssrvs.NET_DFS_ENTRY_ID_CONTAINER, force int32) error {
	req := &netrDfsCreateLocalPartitionRequest{
		ServerName:   optWStr(serverName),
		ShareName:    ndr.WSTR(shareName),
		EntryUid:     dtyp.NewGUID(entryUid),
		EntryPrefix:  ndr.WSTR(entryPrefix),
		ShortName:    ndr.WSTR(shortName),
		RelationInfo: relationInfo,
		Force:        force,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrDfsCreateLocalPartition: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return fmt.Errorf("NetrDfsCreateLocalPartition failed: %s", srvsvc.StatusString(status))
	}
	return nil
}

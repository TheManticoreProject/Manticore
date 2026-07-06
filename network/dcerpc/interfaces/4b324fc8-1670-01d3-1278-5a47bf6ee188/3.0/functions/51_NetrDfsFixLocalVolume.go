package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mssrvs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// netrDfsFixLocalVolumeRequest is the [in] parameter set of NetrDfsFixLocalVolume: the
// [unique] server name, the (ref) volume name, the entry and service types, the (ref)
// storage id, the inline entry GUID (a single [in] pointer in the IDL), the (ref) entry
// prefix, the inline NET_DFS_ENTRY_ID_CONTAINER relation info (a single [in] ref pointer),
// and the create disposition.
type netrDfsFixLocalVolumeRequest struct {
	ServerName        *ndr.WSTR `ndr:"unique"`
	VolumeName        ndr.WSTR
	EntryType         ndr.DWORD
	ServiceType       ndr.DWORD
	StgId             ndr.WSTR
	EntryUid          msdtyp.GUID
	EntryPrefix       ndr.WSTR
	RelationInfo      mssrvs.NET_DFS_ENTRY_ID_CONTAINER
	CreateDisposition ndr.DWORD
}

func (*netrDfsFixLocalVolumeRequest) Opnum() uint16 { return srvsvc.OpnumNetrDfsFixLocalVolume }

// NetrDfsFixLocalVolume calls NetrDfsFixLocalVolume (opnum 51), repairing the local DFS
// volume state ([MS-SRVS] 3.1.4.51).
func NetrDfsFixLocalVolume(rpc ndr.Invoker, serverName, volumeName string, entryType, serviceType uint32, stgId string, entryUid guid.GUID, entryPrefix string, relationInfo mssrvs.NET_DFS_ENTRY_ID_CONTAINER, createDisposition uint32) error {
	req := &netrDfsFixLocalVolumeRequest{
		ServerName:        optWStr(serverName),
		VolumeName:        ndr.WSTR(volumeName),
		EntryType:         ndr.DWORD(entryType),
		ServiceType:       ndr.DWORD(serviceType),
		StgId:             ndr.WSTR(stgId),
		EntryUid:          msdtyp.NewGUID(entryUid),
		EntryPrefix:       ndr.WSTR(entryPrefix),
		RelationInfo:      relationInfo,
		CreateDisposition: ndr.DWORD(createDisposition),
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrDfsFixLocalVolume: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return fmt.Errorf("NetrDfsFixLocalVolume failed: %s", srvsvc.StatusString(status))
	}
	return nil
}

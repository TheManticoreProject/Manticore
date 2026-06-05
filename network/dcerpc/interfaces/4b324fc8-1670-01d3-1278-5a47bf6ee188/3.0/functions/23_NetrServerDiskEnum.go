package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrServerDiskEnumRequest is the [in]/[in,out] parameter set of NetrServerDiskEnum:
// the [unique] server name, the info level, the [in,out] disk-enumeration container,
// the byte budget, and the optional [in,out,unique] resume handle.
type netrServerDiskEnumRequest struct {
	ServerName            *ndr.WSTR `ndr:"unique"`
	Level                 ndr.DWORD
	DiskInfoStruct        structures.DISK_ENUM_CONTAINER
	PreferedMaximumLength ndr.DWORD
	ResumeHandle          *ndr.DWORD `ndr:"unique"`
}

func (*netrServerDiskEnumRequest) Opnum() uint16 { return srvsvc.OpnumNetrServerDiskEnum }

// netrServerDiskEnumResponse is the reply: the updated [in,out] container, the [out]
// total entry count, the updated [in,out,unique] resume handle, and the NET_API_STATUS.
type netrServerDiskEnumResponse struct {
	DiskInfoStruct structures.DISK_ENUM_CONTAINER
	TotalEntries   ndr.DWORD
	ResumeHandle   *ndr.DWORD `ndr:"unique"`
	Status         ndr.DWORD  `ndr:"retval"`
}

// NetrServerDiskEnum calls NetrServerDiskEnum (opnum 23), retrieving a list of disk
// drives on the server ([MS-SRVS] 3.1.4.19). The enumeration is stateful: pass the
// returned resume handle back to continue, starting from 0. ERROR_MORE_DATA indicates
// more pages remain and is not treated as an error.
func NetrServerDiskEnum(rpc ndr.Invoker, serverName string, level uint32, diskInfo structures.DISK_ENUM_CONTAINER, preferedMaximumLength, resumeHandle uint32) (structures.DISK_ENUM_CONTAINER, uint32, uint32, error) {
	resume := ndr.DWORD(resumeHandle)
	req := &netrServerDiskEnumRequest{
		ServerName:            optWStr(serverName),
		Level:                 ndr.DWORD(level),
		DiskInfoStruct:        diskInfo,
		PreferedMaximumLength: ndr.DWORD(preferedMaximumLength),
		ResumeHandle:          &resume,
	}
	var resp netrServerDiskEnumResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.DISK_ENUM_CONTAINER{}, 0, resumeHandle, fmt.Errorf("NetrServerDiskEnum: %w", err)
	}
	var outResume uint32
	if resp.ResumeHandle != nil {
		outResume = uint32(*resp.ResumeHandle)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return resp.DiskInfoStruct, uint32(resp.TotalEntries), outResume, fmt.Errorf("NetrServerDiskEnum failed: %s", srvsvc.StatusString(status))
	}
	return resp.DiskInfoStruct, uint32(resp.TotalEntries), outResume, nil
}

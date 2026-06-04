package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
)

// netrpGetFileSecurityRequest is the [in] parameter set of NetrpGetFileSecurity: the
// [unique] server name, the [unique] share name, the (ref) file name, and the requested
// SECURITY_INFORMATION bits.
type netrpGetFileSecurityRequest struct {
	ServerName           *ndr.WSTR `ndr:"unique"`
	ShareName            *ndr.WSTR `ndr:"unique"`
	LpFileName           ndr.WSTR
	RequestedInformation ndr.DWORD
}

func (*netrpGetFileSecurityRequest) Opnum() uint16 { return srvsvc.OpnumNetrpGetFileSecurity }

// netrpGetFileSecurityResponse is the reply: the [out] PADT_SECURITY_DESCRIPTOR (a double
// pointer in the IDL, modelled as a [unique] pointer) and the return value.
type netrpGetFileSecurityResponse struct {
	SecurityDescriptor *structures.ADT_SECURITY_DESCRIPTOR `ndr:"unique"`
	Status             ndr.DWORD                           `ndr:"retval"`
}

// NetrpGetFileSecurity calls NetrpGetFileSecurity (opnum 39), retrieving the security
// descriptor of a file or directory on a share ([MS-SRVS] 3.1.4.40).
func NetrpGetFileSecurity(rpc *client.Client, serverName, shareName, lpFileName string, requestedInformation uint32) (*structures.ADT_SECURITY_DESCRIPTOR, error) {
	req := &netrpGetFileSecurityRequest{
		ServerName:           optWStr(serverName),
		ShareName:            optWStr(shareName),
		LpFileName:           ndr.WSTR(lpFileName),
		RequestedInformation: ndr.DWORD(requestedInformation),
	}
	var resp netrpGetFileSecurityResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("NetrpGetFileSecurity: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return resp.SecurityDescriptor, fmt.Errorf("NetrpGetFileSecurity failed: %s", srvsvc.StatusString(status))
	}
	return resp.SecurityDescriptor, nil
}

package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrConnectRequest carries the [in] parameters of SamrConnect: the [unique]
// server name (PSAMPR_SERVER_NAME, NULL for the local server) and the desired
// access mask for the returned server handle.
type samrConnectRequest struct {
	ServerName    *ndr.WSTR `ndr:"unique"`
	DesiredAccess ndr.DWORD
}

func (*samrConnectRequest) Opnum() uint16 { return samr.OpnumSamrConnect }

// SamrConnect calls SamrConnect (opnum 0), obtaining a handle to a server object
// ([MS-SAMR] 3.1.5.1.4).
func SamrConnect(rpc *client.Client, serverName string, desiredAccess uint32) (structures.SAMPR_HANDLE, error) {
	req := &samrConnectRequest{
		ServerName:    optWStr(serverName),
		DesiredAccess: ndr.DWORD(desiredAccess),
	}
	var resp openHandleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.SAMPR_HANDLE{}, fmt.Errorf("SamrConnect: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Handle, fmt.Errorf("SamrConnect failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}

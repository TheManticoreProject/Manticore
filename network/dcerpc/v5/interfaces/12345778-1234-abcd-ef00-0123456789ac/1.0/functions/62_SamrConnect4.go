package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrConnect4Request carries the [in] parameters of SamrConnect4: the [unique]
// server name (NULL for the local server), the client revision, and the desired
// access mask.
type samrConnect4Request struct {
	ServerName     *ndr.WSTR `ndr:"unique"`
	ClientRevision ndr.DWORD
	DesiredAccess  ndr.DWORD
}

func (*samrConnect4Request) Opnum() uint16 { return samr.OpnumSamrConnect4 }

// SamrConnect4 calls SamrConnect4 (opnum 62), obtaining a handle to a server
// object and advertising the client revision ([MS-SAMR] 3.1.5.1.2).
func SamrConnect4(rpc *client.Client, serverName string, clientRevision uint32, desiredAccess uint32) (structures.SAMPR_HANDLE, error) {
	req := &samrConnect4Request{
		ServerName:     optWStr(serverName),
		ClientRevision: ndr.DWORD(clientRevision),
		DesiredAccess:  ndr.DWORD(desiredAccess),
	}
	var resp openHandleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.SAMPR_HANDLE{}, fmt.Errorf("SamrConnect4: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Handle, fmt.Errorf("SamrConnect4 failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}

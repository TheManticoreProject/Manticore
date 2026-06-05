package functions

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// samrConnect2Request carries the [in] parameters of SamrConnect2: the [unique]
// server name (NULL for the local server) and the desired access mask.
type samrConnect2Request struct {
	ServerName    *ndr.WSTR `ndr:"unique"`
	DesiredAccess ndr.DWORD
}

func (*samrConnect2Request) Opnum() uint16 { return samr.OpnumSamrConnect2 }

// SamrConnect2 calls SamrConnect2 (opnum 57), obtaining a handle to a server
// object ([MS-SAMR] 3.1.5.1.3).
func SamrConnect2(rpc ndr.Invoker, serverName string, desiredAccess uint32) (structures.SAMPR_HANDLE, error) {
	req := &samrConnect2Request{
		ServerName:    optWStr(serverName),
		DesiredAccess: ndr.DWORD(desiredAccess),
	}
	var resp openHandleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.SAMPR_HANDLE{}, fmt.Errorf("SamrConnect2: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Handle, fmt.Errorf("SamrConnect2 failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}

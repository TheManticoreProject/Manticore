package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrConnect5Request carries the [in] parameters of SamrConnect5: the [unique]
// server name (NULL for the local server), the desired access mask, the requested
// input version, and the [unique, switch_is(InVersion)] revision-info union.
type samrConnect5Request struct {
	ServerName     *ndr.WSTR `ndr:"unique"`
	DesiredAccess  ndr.DWORD
	InVersion      ndr.DWORD
	InRevisionInfo *structures.SAMPR_REVISION_INFO `ndr:"unique"`
}

func (*samrConnect5Request) Opnum() uint16 { return samr.OpnumSamrConnect5 }

// samrConnect5Response carries the [out] negotiated output version, the [unique,
// switch_is(*OutVersion)] revision-info union, the server handle, and the NTSTATUS.
type samrConnect5Response struct {
	OutVersion      ndr.DWORD
	OutRevisionInfo *structures.SAMPR_REVISION_INFO `ndr:"unique"`
	ServerHandle    structures.SAMPR_HANDLE
	Status          ndr.DWORD `ndr:"retval"`
}

// SamrConnect5 calls SamrConnect5 (opnum 64), obtaining a handle to a server
// object while negotiating a revision ([MS-SAMR] 3.1.5.1.1). This convenience
// signature requests input version 1 with a V1 revision info ({Revision: 3}).
func SamrConnect5(rpc *client.Client, serverName string, desiredAccess uint32) (structures.SAMPR_HANDLE, uint32, *structures.SAMPR_REVISION_INFO, error) {
	const inVersion = 1
	req := &samrConnect5Request{
		ServerName:    optWStr(serverName),
		DesiredAccess: ndr.DWORD(desiredAccess),
		InVersion:     ndr.DWORD(inVersion),
		InRevisionInfo: &structures.SAMPR_REVISION_INFO{
			Tag: ndr.DWORD(inVersion),
			V1:  structures.SAMPR_REVISION_INFO_V1{Revision: 3},
		},
	}
	var resp samrConnect5Response
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.SAMPR_HANDLE{}, 0, nil, fmt.Errorf("SamrConnect5: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.ServerHandle, uint32(resp.OutVersion), resp.OutRevisionInfo, fmt.Errorf("SamrConnect5 failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.ServerHandle, uint32(resp.OutVersion), resp.OutRevisionInfo, nil
}

package functions

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// samrConnect5Request carries the [in] parameters of SamrConnect5: the [unique]
// server name (NULL for the local server), the desired access mask, the requested
// input version, and the switch_is(InVersion) revision-info union.
//
// InRevisionInfo is the IDL's [switch_is(InVersion)] SAMPR_REVISION_INFO* parameter.
// A switch_is union argument is transmitted inline (its own discriminant followed by
// the selected arm), NOT wrapped behind a pointer referent — emitting a referent id
// here is rejected by the server as nca_s_fault_ndr.
type samrConnect5Request struct {
	ServerName     *ndr.WSTR `ndr:"unique"`
	DesiredAccess  ndr.DWORD
	InVersion      ndr.DWORD
	InRevisionInfo structures.SAMPR_REVISION_INFO
}

func (*samrConnect5Request) Opnum() uint16 { return samr.OpnumSamrConnect5 }

// samrConnect5Response carries the [out] negotiated output version, the
// switch_is(*OutVersion) revision-info union (transmitted inline), the server handle,
// and the NTSTATUS.
type samrConnect5Response struct {
	OutVersion      ndr.DWORD
	OutRevisionInfo structures.SAMPR_REVISION_INFO
	ServerHandle    structures.SAMPR_HANDLE
	Status          ndr.DWORD `ndr:"retval"`
}

// SamrConnect5 calls SamrConnect5 (opnum 64), obtaining a handle to a server
// object while negotiating a revision ([MS-SAMR] 3.1.5.1.1). This convenience
// signature requests input version 1 with a V1 revision info ({Revision: 3}).
func SamrConnect5(rpc ndr.Invoker, serverName string, desiredAccess uint32) (structures.SAMPR_HANDLE, uint32, *structures.SAMPR_REVISION_INFO, error) {
	const inVersion = 1
	req := &samrConnect5Request{
		ServerName:    optWStr(serverName),
		DesiredAccess: ndr.DWORD(desiredAccess),
		InVersion:     ndr.DWORD(inVersion),
		InRevisionInfo: structures.SAMPR_REVISION_INFO{
			Tag: ndr.DWORD(inVersion),
			V1:  structures.SAMPR_REVISION_INFO_V1{Revision: 3},
		},
	}
	var resp samrConnect5Response
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.SAMPR_HANDLE{}, 0, nil, fmt.Errorf("SamrConnect5: %w", err)
	}
	outInfo := resp.OutRevisionInfo
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.ServerHandle, uint32(resp.OutVersion), &outInfo, fmt.Errorf("SamrConnect5 failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.ServerHandle, uint32(resp.OutVersion), &outInfo, nil
}

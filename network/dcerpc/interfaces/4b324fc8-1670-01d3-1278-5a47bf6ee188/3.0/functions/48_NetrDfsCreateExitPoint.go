package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// netrDfsCreateExitPointRequest is the [in] parameter set of NetrDfsCreateExitPoint: the
// [unique] server name, the inline GUID (a single [in] pointer in the IDL), the (ref)
// prefix, the exit-point type, and the output prefix length sizing ShortPrefix.
type netrDfsCreateExitPointRequest struct {
	ServerName     *ndr.WSTR `ndr:"unique"`
	Uid            dtyp.GUID
	Prefix         ndr.WSTR
	Type           ndr.DWORD
	ShortPrefixLen ndr.DWORD
}

func (*netrDfsCreateExitPointRequest) Opnum() uint16 { return srvsvc.OpnumNetrDfsCreateExitPoint }

// netrDfsCreateExitPointResponse is the reply: the [out, size_is(ShortPrefixLen)] WCHAR
// short prefix and the NET_API_STATUS return value.
type netrDfsCreateExitPointResponse struct {
	ShortPrefix []uint16  `ndr:"ref,size_is=ShortPrefixLen"`
	Status      ndr.DWORD `ndr:"retval"`
}

// NetrDfsCreateExitPoint calls NetrDfsCreateExitPoint (opnum 48), creating a DFS exit
// point ([MS-SRVS] 3.1.4.48).
func NetrDfsCreateExitPoint(rpc ndr.Invoker, serverName string, uid guid.GUID, prefix string, typ, shortPrefixLen uint32) ([]uint16, error) {
	req := &netrDfsCreateExitPointRequest{
		ServerName:     optWStr(serverName),
		Uid:            dtyp.NewGUID(uid),
		Prefix:         ndr.WSTR(prefix),
		Type:           ndr.DWORD(typ),
		ShortPrefixLen: ndr.DWORD(shortPrefixLen),
	}
	var resp netrDfsCreateExitPointResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("NetrDfsCreateExitPoint: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return resp.ShortPrefix, fmt.Errorf("NetrDfsCreateExitPoint failed: %s", srvsvc.StatusString(status))
	}
	return resp.ShortPrefix, nil
}

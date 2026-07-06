package functions

// IDL source: [MS-SRVS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/77aacc74-f8f9-4b46-b2d8-bfe04a7d9c44
// A fetched copy is kept at ms-srvs.idl in the interface directory.

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrShareCheckRequest is the [in] parameter set of NetrShareCheck: the optional server
// name and the [in,string] (ref) device name to test.
type netrShareCheckRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Device     ndr.WSTR
}

func (*netrShareCheckRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrShareCheck
}

// netrShareCheckResponse is the reply: the share type ([out] DWORD) and the NET_API_STATUS
// return value.
type netrShareCheckResponse struct {
	Type   ndr.DWORD
	Status ndr.DWORD `ndr:"retval"`
}

// NetrShareCheck calls NetrShareCheck (opnum 20), checking whether a device is shared
// ([MS-SRVS] 3.1.4.14).
func NetrShareCheck(rpc ndr.Invoker, serverName string, device string) (ndr.DWORD, error) {
	req := &netrShareCheckRequest{
		ServerName: optWStr(serverName),
		Device:     ndr.WSTR(device),
	}
	var resp netrShareCheckResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return 0, fmt.Errorf("NetrShareCheck: %w", err)
	}
	if uint32(resp.Status) != srvsvc.NERR_Success && uint32(resp.Status) != srvsvc.ERROR_MORE_DATA {
		return resp.Type, fmt.Errorf("NetrShareCheck failed: %s", srvsvc.StatusString(uint32(resp.Status)))
	}
	return resp.Type, nil
}

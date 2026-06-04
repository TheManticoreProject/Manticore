package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
)

// netprPathCanonicalizeRequest is the [in] parameter set of NetprPathCanonicalize: the
// [unique] server name, the (ref) path name, the output buffer length sizing Outbuf, the
// (ref) prefix, the [in,out] path type, and the type flags.
type netprPathCanonicalizeRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	PathName   ndr.WSTR
	OutbufLen  ndr.DWORD
	Prefix     ndr.WSTR
	PathType   ndr.DWORD
	Flags      ndr.DWORD
}

func (*netprPathCanonicalizeRequest) Opnum() uint16 { return srvsvc.OpnumNetprPathCanonicalize }

// netprPathCanonicalizeResponse is the reply: the [out, size_is(OutbufLen)] byte buffer,
// the [in,out] path type, and the NET_API_STATUS return value.
type netprPathCanonicalizeResponse struct {
	Outbuf   []byte `ndr:"ref,size_is=OutbufLen"`
	PathType ndr.DWORD
	Status   ndr.DWORD `ndr:"retval"`
}

// NetprPathCanonicalize calls NetprPathCanonicalize (opnum 31), converting a path name to
// its canonical form ([MS-SRVS] 3.1.4.27).
func NetprPathCanonicalize(rpc *client.Client, serverName, pathName string, outbufLen uint32, prefix string, pathType, flags uint32) ([]byte, uint32, error) {
	req := &netprPathCanonicalizeRequest{
		ServerName: optWStr(serverName),
		PathName:   ndr.WSTR(pathName),
		OutbufLen:  ndr.DWORD(outbufLen),
		Prefix:     ndr.WSTR(prefix),
		PathType:   ndr.DWORD(pathType),
		Flags:      ndr.DWORD(flags),
	}
	var resp netprPathCanonicalizeResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, 0, fmt.Errorf("NetprPathCanonicalize: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return resp.Outbuf, uint32(resp.PathType), fmt.Errorf("NetprPathCanonicalize failed: %s", srvsvc.StatusString(status))
	}
	return resp.Outbuf, uint32(resp.PathType), nil
}

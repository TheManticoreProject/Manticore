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

// netprNameCanonicalizeRequest is the [in] parameter set of NetprNameCanonicalize: the
// [unique] server name, the (ref) name, the output buffer length sizing Outbuf, the name
// type, and the flags.
type netprNameCanonicalizeRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Name       ndr.WSTR
	OutbufLen  ndr.DWORD
	NameType   ndr.DWORD
	Flags      ndr.DWORD
}

func (*netprNameCanonicalizeRequest) Opnum() uint16 { return srvsvc.OpnumNetprNameCanonicalize }

// netprNameCanonicalizeResponse is the reply: the [out, size_is(OutbufLen)] WCHAR buffer
// and the NET_API_STATUS return value.
type netprNameCanonicalizeResponse struct {
	Outbuf []uint16  `ndr:"ref,size_is=OutbufLen"`
	Status ndr.DWORD `ndr:"retval"`
}

// NetprNameCanonicalize calls NetprNameCanonicalize (opnum 34), converting a name to its
// canonical form ([MS-SRVS] 3.1.4.31).
func NetprNameCanonicalize(rpc ndr.Invoker, serverName, name string, outbufLen, nameType, flags uint32) ([]uint16, error) {
	req := &netprNameCanonicalizeRequest{
		ServerName: optWStr(serverName),
		Name:       ndr.WSTR(name),
		OutbufLen:  ndr.DWORD(outbufLen),
		NameType:   ndr.DWORD(nameType),
		Flags:      ndr.DWORD(flags),
	}
	var resp netprNameCanonicalizeResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("NetprNameCanonicalize: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return resp.Outbuf, fmt.Errorf("NetprNameCanonicalize failed: %s", srvsvc.StatusString(status))
	}
	return resp.Outbuf, nil
}

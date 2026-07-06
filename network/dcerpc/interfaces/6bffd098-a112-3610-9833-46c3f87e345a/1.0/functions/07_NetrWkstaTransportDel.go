package functions

// IDL source: [MS-WKST] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-wkst/9fdbc753-0397-4236-bbfc-a380f9d23789
// A fetched copy is kept at ms-wkst.idl in the interface directory.

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrWkstaTransportDelRequest carries the [in] parameters of NetrWkstaTransportDel.
type netrWkstaTransportDelRequest struct {
	ServerName    *ndr.WSTR `ndr:"unique"`
	TransportName *ndr.WSTR `ndr:"unique"`
	ForceLevel    ndr.DWORD
}

func (*netrWkstaTransportDelRequest) Opnum() uint16 { return wkssvc.OpnumNetrWkstaTransportDel }

// netrWkstaTransportDelResponse carries the [out] parameters and return value of NetrWkstaTransportDel.
type netrWkstaTransportDelResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrWkstaTransportDel calls NetrWkstaTransportDel (opnum 7) ([MS-WKST] 3.2.4).
func NetrWkstaTransportDel(rpc ndr.Invoker, serverName *ndr.WSTR, transportName *ndr.WSTR, forceLevel ndr.DWORD) (err error) {
	req := &netrWkstaTransportDelRequest{
		ServerName:    serverName,
		TransportName: transportName,
		ForceLevel:    forceLevel,
	}
	var resp netrWkstaTransportDelResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrWkstaTransportDel: %w", err)
		return
	}
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrWkstaTransportDel failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}

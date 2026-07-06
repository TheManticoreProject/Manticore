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

// netrUseDelRequest carries the [in] parameters of NetrUseDel.
type netrUseDelRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	UseName    ndr.WSTR
	ForceLevel ndr.DWORD
}

func (*netrUseDelRequest) Opnum() uint16 { return wkssvc.OpnumNetrUseDel }

// netrUseDelResponse carries the [out] parameters and return value of NetrUseDel.
type netrUseDelResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrUseDel calls NetrUseDel (opnum 10) ([MS-WKST] 3.2.4).
func NetrUseDel(rpc ndr.Invoker, serverName *ndr.WSTR, useName ndr.WSTR, forceLevel ndr.DWORD) (err error) {
	req := &netrUseDelRequest{
		ServerName: serverName,
		UseName:    useName,
		ForceLevel: forceLevel,
	}
	var resp netrUseDelResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrUseDel: %w", err)
		return
	}
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrUseDel failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}

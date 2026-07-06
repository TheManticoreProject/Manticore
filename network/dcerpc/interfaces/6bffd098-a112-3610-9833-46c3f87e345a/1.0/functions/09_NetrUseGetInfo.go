package functions

// IDL source: [MS-WKST] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-wkst/9fdbc753-0397-4236-bbfc-a380f9d23789
// A fetched copy is kept at ms-wkst.idl in the interface directory.

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrUseGetInfoRequest carries the [in] parameters of NetrUseGetInfo.
type netrUseGetInfoRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	UseName    ndr.WSTR
	Level      ndr.DWORD
}

func (*netrUseGetInfoRequest) Opnum() uint16 { return wkssvc.OpnumNetrUseGetInfo }

// netrUseGetInfoResponse carries the [out] parameters and return value of NetrUseGetInfo.
type netrUseGetInfoResponse struct {
	InfoStruct mswkst.USE_INFO
	Status     ndr.DWORD `ndr:"retval"`
}

// NetrUseGetInfo calls NetrUseGetInfo (opnum 9) ([MS-WKST] 3.2.4).
func NetrUseGetInfo(rpc ndr.Invoker, serverName *ndr.WSTR, useName ndr.WSTR, level ndr.DWORD) (InfoStruct mswkst.USE_INFO, err error) {
	req := &netrUseGetInfoRequest{
		ServerName: serverName,
		UseName:    useName,
		Level:      level,
	}
	var resp netrUseGetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrUseGetInfo: %w", err)
		return
	}
	InfoStruct = resp.InfoStruct
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrUseGetInfo failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}

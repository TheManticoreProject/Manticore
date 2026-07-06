package functions

// IDL source: [MS-DFSNM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dfsnm/b471e023-618d-4c48-877f-f30c3005320c
// A fetched copy is kept at ms-dfsnm.idl in the interface directory.

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrDfsAddStdRootForcedRequest carries the [in] parameters of NetrDfsAddStdRootForced.
type netrDfsAddStdRootForcedRequest struct {
	ServerName ndr.WSTR
	RootShare  ndr.WSTR
	Comment    ndr.WSTR
	Share      ndr.WSTR
}

func (*netrDfsAddStdRootForcedRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsAddStdRootForced }

// netrDfsAddStdRootForcedResponse carries the [out] parameters and return value of NetrDfsAddStdRootForced.
type netrDfsAddStdRootForcedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrDfsAddStdRootForced calls NetrDfsAddStdRootForced (opnum 15) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsAddStdRootForced(rpc ndr.Invoker, serverName ndr.WSTR, rootShare ndr.WSTR, comment ndr.WSTR, share ndr.WSTR) (err error) {
	req := &netrDfsAddStdRootForcedRequest{
		ServerName: serverName,
		RootShare:  rootShare,
		Comment:    comment,
		Share:      share,
	}
	var resp netrDfsAddStdRootForcedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsAddStdRootForced: %w", err)
		return
	}
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsAddStdRootForced failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}

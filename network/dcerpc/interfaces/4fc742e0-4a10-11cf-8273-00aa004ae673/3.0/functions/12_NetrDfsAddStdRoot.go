package functions

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrDfsAddStdRootRequest carries the [in] parameters of NetrDfsAddStdRoot.
type netrDfsAddStdRootRequest struct {
	ServerName ndr.WSTR
	RootShare  ndr.WSTR
	Comment    ndr.WSTR
	ApiFlags   ndr.DWORD
}

func (*netrDfsAddStdRootRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsAddStdRoot }

// netrDfsAddStdRootResponse carries the [out] parameters and return value of NetrDfsAddStdRoot.
type netrDfsAddStdRootResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrDfsAddStdRoot calls NetrDfsAddStdRoot (opnum 12) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsAddStdRoot(rpc ndr.Invoker, serverName ndr.WSTR, rootShare ndr.WSTR, comment ndr.WSTR, apiFlags ndr.DWORD) (err error) {
	req := &netrDfsAddStdRootRequest{
		ServerName: serverName,
		RootShare:  rootShare,
		Comment:    comment,
		ApiFlags:   apiFlags,
	}
	var resp netrDfsAddStdRootResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsAddStdRoot: %w", err)
		return
	}
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsAddStdRoot failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}

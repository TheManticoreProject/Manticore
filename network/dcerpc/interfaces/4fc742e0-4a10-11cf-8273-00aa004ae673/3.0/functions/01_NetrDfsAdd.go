package functions

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrDfsAddRequest carries the [in] parameters of NetrDfsAdd.
type netrDfsAddRequest struct {
	DfsEntryPath ndr.WSTR
	ServerName   ndr.WSTR
	ShareName    *ndr.WSTR `ndr:"unique"`
	Comment      *ndr.WSTR `ndr:"unique"`
	Flags        ndr.DWORD
}

func (*netrDfsAddRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsAdd }

// netrDfsAddResponse carries the [out] parameters and return value of NetrDfsAdd.
type netrDfsAddResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrDfsAdd calls NetrDfsAdd (opnum 1) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsAdd(rpc ndr.Invoker, dfsEntryPath ndr.WSTR, serverName ndr.WSTR, shareName *ndr.WSTR, comment *ndr.WSTR, flags ndr.DWORD) (err error) {
	req := &netrDfsAddRequest{
		DfsEntryPath: dfsEntryPath,
		ServerName:   serverName,
		ShareName:    shareName,
		Comment:      comment,
		Flags:        flags,
	}
	var resp netrDfsAddResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsAdd: %w", err)
		return
	}
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsAdd failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrDfsFlushFtTableRequest carries the [in] parameters of NetrDfsFlushFtTable.
type netrDfsFlushFtTableRequest struct {
	DcName       ndr.WSTR
	WszFtDfsName ndr.WSTR
}

func (*netrDfsFlushFtTableRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsFlushFtTable }

// netrDfsFlushFtTableResponse carries the [out] parameters and return value of NetrDfsFlushFtTable.
type netrDfsFlushFtTableResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrDfsFlushFtTable calls NetrDfsFlushFtTable (opnum 18) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsFlushFtTable(rpc ndr.Invoker, dcName ndr.WSTR, wszFtDfsName ndr.WSTR) (err error) {
	req := &netrDfsFlushFtTableRequest{
		DcName:       dcName,
		WszFtDfsName: wszFtDfsName,
	}
	var resp netrDfsFlushFtTableResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsFlushFtTable: %w", err)
		return
	}
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsFlushFtTable failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}

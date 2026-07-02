package functions

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrDfsGetDcAddressRequest carries the [in] parameters of NetrDfsGetDcAddress.
type netrDfsGetDcAddressRequest struct {
	ServerName ndr.WSTR
	DcName     ndr.WSTR
	IsRoot     bool
	Timeout    ndr.DWORD
}

func (*netrDfsGetDcAddressRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsGetDcAddress }

// netrDfsGetDcAddressResponse carries the [out] parameters and return value of NetrDfsGetDcAddress.
type netrDfsGetDcAddressResponse struct {
	DcName  ndr.WSTR
	IsRoot  bool
	Timeout ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// NetrDfsGetDcAddress calls NetrDfsGetDcAddress (opnum 16) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsGetDcAddress(rpc ndr.Invoker, serverName ndr.WSTR, dcName ndr.WSTR, isRoot bool, timeout ndr.DWORD) (DcName ndr.WSTR, IsRoot bool, Timeout ndr.DWORD, err error) {
	req := &netrDfsGetDcAddressRequest{
		ServerName: serverName,
		DcName:     dcName,
		IsRoot:     isRoot,
		Timeout:    timeout,
	}
	var resp netrDfsGetDcAddressResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsGetDcAddress: %w", err)
		return
	}
	DcName = resp.DcName
	IsRoot = resp.IsRoot
	Timeout = resp.Timeout
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsGetDcAddress failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}

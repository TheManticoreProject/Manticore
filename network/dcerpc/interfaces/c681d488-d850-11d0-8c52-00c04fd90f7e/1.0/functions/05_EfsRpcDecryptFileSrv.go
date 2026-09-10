package functions

// IDL source: [MS-EFSR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-efsr/4a25b8e1-fd90-41b6-9301-62ed71334436
// A fetched copy is kept at ms-efsr.idl in the interface directory.

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// efsRpcDecryptFileSrvRequest carries the [in] parameters of EfsRpcDecryptFileSrv.
type efsRpcDecryptFileSrvRequest struct {
	FileName ndr.WSTR
	OpenFlag ndr.DWORD
}

func (*efsRpcDecryptFileSrvRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcDecryptFileSrv }

// efsRpcDecryptFileSrvResponse carries the [out] parameters and return value of EfsRpcDecryptFileSrv.
type efsRpcDecryptFileSrvResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EfsRpcDecryptFileSrv calls EfsRpcDecryptFileSrv (opnum 5) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcDecryptFileSrv(rpc ndr.Invoker, fileName ndr.WSTR, openFlag ndr.DWORD) (err error) {
	req := &efsRpcDecryptFileSrvRequest{
		FileName: fileName,
		OpenFlag: openFlag,
	}
	var resp efsRpcDecryptFileSrvResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcDecryptFileSrv: %w", err)
		return
	}
	if win32.WIN32_ERROR(resp.Status) != win32.ERROR_SUCCESS {
		err = fmt.Errorf("EfsRpcDecryptFileSrv failed: %s", win32.WIN32_ERROR(resp.Status).String())
	}
	return
}

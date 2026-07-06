package functions

// IDL source: [MS-EFSR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-efsr/4a25b8e1-fd90-41b6-9301-62ed71334436
// A fetched copy is kept at ms-efsr.idl in the interface directory.

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msefsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-efsr"
)

// efsRpcDuplicateEncryptionInfoFileRequest carries the [in] parameters of EfsRpcDuplicateEncryptionInfoFile.
type efsRpcDuplicateEncryptionInfoFileRequest struct {
	SrcFileName           ndr.WSTR
	DestFileName          ndr.WSTR
	DwCreationDisposition ndr.DWORD
	DwAttributes          ndr.DWORD
	RelativeSD            *msefsr.EFS_RPC_BLOB `ndr:"unique"`
	BInheritHandle        ndr.BOOL
}

func (*efsRpcDuplicateEncryptionInfoFileRequest) Opnum() uint16 {
	return efsrpc.OpnumEfsRpcDuplicateEncryptionInfoFile
}

// efsRpcDuplicateEncryptionInfoFileResponse carries the [out] parameters and return value of EfsRpcDuplicateEncryptionInfoFile.
type efsRpcDuplicateEncryptionInfoFileResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EfsRpcDuplicateEncryptionInfoFile calls EfsRpcDuplicateEncryptionInfoFile (opnum 13) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcDuplicateEncryptionInfoFile(rpc ndr.Invoker, srcFileName ndr.WSTR, destFileName ndr.WSTR, dwCreationDisposition ndr.DWORD, dwAttributes ndr.DWORD, relativeSD *msefsr.EFS_RPC_BLOB, bInheritHandle ndr.BOOL) (err error) {
	req := &efsRpcDuplicateEncryptionInfoFileRequest{
		SrcFileName:           srcFileName,
		DestFileName:          destFileName,
		DwCreationDisposition: dwCreationDisposition,
		DwAttributes:          dwAttributes,
		RelativeSD:            relativeSD,
		BInheritHandle:        bInheritHandle,
	}
	var resp efsRpcDuplicateEncryptionInfoFileResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcDuplicateEncryptionInfoFile: %w", err)
		return
	}
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcDuplicateEncryptionInfoFile failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}

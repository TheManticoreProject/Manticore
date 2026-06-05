package functions

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// efsRpcDuplicateEncryptionInfoFileRequest carries the [in] parameters of EfsRpcDuplicateEncryptionInfoFile.
type efsRpcDuplicateEncryptionInfoFileRequest struct {
	SrcFileName           *ndr.WSTR `ndr:"unique"`
	DestFileName          *ndr.WSTR `ndr:"unique"`
	DwCreationDisposition ndr.DWORD
	DwAttributes          ndr.DWORD
	RelativeSD            *structures.EFS_RPC_BLOB `ndr:"unique"`
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
func EfsRpcDuplicateEncryptionInfoFile(rpc ndr.Invoker, srcFileName *ndr.WSTR, destFileName *ndr.WSTR, dwCreationDisposition ndr.DWORD, dwAttributes ndr.DWORD, relativeSD *structures.EFS_RPC_BLOB, bInheritHandle ndr.BOOL) (err error) {
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

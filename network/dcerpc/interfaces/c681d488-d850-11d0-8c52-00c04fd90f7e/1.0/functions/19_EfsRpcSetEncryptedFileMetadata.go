package functions

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// efsRpcSetEncryptedFileMetadataRequest carries the [in] parameters of EfsRpcSetEncryptedFileMetadata.
type efsRpcSetEncryptedFileMetadataRequest struct {
	FileName         *ndr.WSTR                `ndr:"unique"`
	OldEfsStreamBlob *structures.EFS_RPC_BLOB `ndr:"unique"`
	NewEfsStreamBlob structures.EFS_RPC_BLOB
	NewEfsSignature  *structures.ENCRYPTED_FILE_METADATA_SIGNATURE `ndr:"unique"`
}

func (*efsRpcSetEncryptedFileMetadataRequest) Opnum() uint16 {
	return efsrpc.OpnumEfsRpcSetEncryptedFileMetadata
}

// efsRpcSetEncryptedFileMetadataResponse carries the [out] parameters and return value of EfsRpcSetEncryptedFileMetadata.
type efsRpcSetEncryptedFileMetadataResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EfsRpcSetEncryptedFileMetadata calls EfsRpcSetEncryptedFileMetadata (opnum 19) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcSetEncryptedFileMetadata(rpc ndr.Invoker, fileName *ndr.WSTR, oldEfsStreamBlob *structures.EFS_RPC_BLOB, newEfsStreamBlob structures.EFS_RPC_BLOB, newEfsSignature *structures.ENCRYPTED_FILE_METADATA_SIGNATURE) (err error) {
	req := &efsRpcSetEncryptedFileMetadataRequest{
		FileName:         fileName,
		OldEfsStreamBlob: oldEfsStreamBlob,
		NewEfsStreamBlob: newEfsStreamBlob,
		NewEfsSignature:  newEfsSignature,
	}
	var resp efsRpcSetEncryptedFileMetadataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcSetEncryptedFileMetadata: %w", err)
		return
	}
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcSetEncryptedFileMetadata failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}

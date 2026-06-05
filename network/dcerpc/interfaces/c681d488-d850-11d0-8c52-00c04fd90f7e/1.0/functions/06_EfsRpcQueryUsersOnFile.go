package functions

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// efsRpcQueryUsersOnFileRequest carries the [in] parameters of EfsRpcQueryUsersOnFile.
type efsRpcQueryUsersOnFileRequest struct {
	FileName *ndr.WSTR `ndr:"unique"`
}

func (*efsRpcQueryUsersOnFileRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcQueryUsersOnFile }

// efsRpcQueryUsersOnFileResponse carries the [out] parameters and return value of EfsRpcQueryUsersOnFile.
type efsRpcQueryUsersOnFileResponse struct {
	Users  *structures.ENCRYPTION_CERTIFICATE_HASH_LIST `ndr:"unique"`
	Status ndr.DWORD                                    `ndr:"retval"`
}

// EfsRpcQueryUsersOnFile calls EfsRpcQueryUsersOnFile (opnum 6) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcQueryUsersOnFile(rpc ndr.Invoker, fileName *ndr.WSTR) (Users *structures.ENCRYPTION_CERTIFICATE_HASH_LIST, err error) {
	req := &efsRpcQueryUsersOnFileRequest{
		FileName: fileName,
	}
	var resp efsRpcQueryUsersOnFileResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcQueryUsersOnFile: %w", err)
		return
	}
	Users = resp.Users
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcQueryUsersOnFile failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}

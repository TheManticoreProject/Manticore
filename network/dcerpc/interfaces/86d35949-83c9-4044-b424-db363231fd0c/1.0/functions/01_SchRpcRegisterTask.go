package functions

import (
	"fmt"

	schrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/86d35949-83c9-4044-b424-db363231fd0c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mstsch "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsch"
)

// schRpcRegisterTaskRequest carries the [in] parameters of SchRpcRegisterTask.
type schRpcRegisterTaskRequest struct {
	Path      *ndr.WSTR `ndr:"unique"`
	Xml       ndr.WSTR
	Flags     ndr.DWORD
	Sddl      *ndr.WSTR `ndr:"unique"`
	LogonType ndr.DWORD
	CCreds    ndr.DWORD
	PCreds    []mstsch.TASK_USER_CRED `ndr:"unique,size_is=CCreds"`
}

func (*schRpcRegisterTaskRequest) Opnum() uint16 {
	return schrpc.OpnumSchRpcRegisterTask
}

// schRpcRegisterTaskResponse carries the [out] parameters and return value of SchRpcRegisterTask.
type schRpcRegisterTaskResponse struct {
	PActualPath *ndr.WSTR                   `ndr:"unique"`
	PErrorInfo  *mstsch.TASK_XML_ERROR_INFO `ndr:"unique"`
	Status      ndr.DWORD                   `ndr:"retval"`
}

// SchRpcRegisterTask calls SchRpcRegisterTask (opnum 1) ([MS-TSCH] section 3.2.5.4.2).
func SchRpcRegisterTask(rpc ndr.Invoker, path *ndr.WSTR, xml ndr.WSTR, flags ndr.DWORD, sddl *ndr.WSTR, logonType ndr.DWORD, cCreds ndr.DWORD, pCreds []mstsch.TASK_USER_CRED) (PActualPath *ndr.WSTR, PErrorInfo *mstsch.TASK_XML_ERROR_INFO, err error) {
	req := &schRpcRegisterTaskRequest{
		Path:      path,
		Xml:       xml,
		Flags:     flags,
		Sddl:      sddl,
		LogonType: logonType,
		CCreds:    cCreds,
		PCreds:    pCreds,
	}
	var resp schRpcRegisterTaskResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcRegisterTask: %w", err)
		return
	}
	PActualPath = resp.PActualPath
	PErrorInfo = resp.PErrorInfo
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcRegisterTask failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}

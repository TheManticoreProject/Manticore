package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	frsapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/d049b186-814f-11d1-9a3c-00c04fc9b232/1.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// ntFrsApi_Rpc_IsPathReplicatedRequest carries the [in] parameters of NtFrsApi_Rpc_IsPathReplicated.
type ntFrsApi_Rpc_IsPathReplicatedRequest struct {
	Path                     *ndr.WSTR `ndr:"unique"`
	ReplicaSetTypeOfInterest ndr.DWORD
}

func (*ntFrsApi_Rpc_IsPathReplicatedRequest) Opnum() uint16 {
	return frsapi.OpnumNtFrsApi_Rpc_IsPathReplicated
}

// ntFrsApi_Rpc_IsPathReplicatedResponse carries the [out] parameters and return value of NtFrsApi_Rpc_IsPathReplicated.
type ntFrsApi_Rpc_IsPathReplicatedResponse struct {
	Replicated     ndr.DWORD
	Primary        ndr.DWORD
	Root           ndr.DWORD
	ReplicaSetGuid dtyp.GUID
	Status         ndr.DWORD `ndr:"retval"`
}

// NtFrsApi_Rpc_IsPathReplicated calls NtFrsApi_Rpc_IsPathReplicated (opnum 8) ([MS-FRS1] section 3.2.4.4).
func NtFrsApi_Rpc_IsPathReplicated(rpc ndr.Invoker, path *ndr.WSTR, replicaSetTypeOfInterest ndr.DWORD) (Replicated ndr.DWORD, Primary ndr.DWORD, Root ndr.DWORD, ReplicaSetGuid dtyp.GUID, err error) {
	req := &ntFrsApi_Rpc_IsPathReplicatedRequest{
		Path:                     path,
		ReplicaSetTypeOfInterest: replicaSetTypeOfInterest,
	}
	var resp ntFrsApi_Rpc_IsPathReplicatedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NtFrsApi_Rpc_IsPathReplicated: %w", err)
		return
	}
	Replicated = resp.Replicated
	Primary = resp.Primary
	Root = resp.Root
	ReplicaSetGuid = resp.ReplicaSetGuid
	if uint32(resp.Status) != frsapi.StatusSuccess {
		err = fmt.Errorf("NtFrsApi_Rpc_IsPathReplicated failed: %s", frsapi.StatusString(uint32(resp.Status)))
	}
	return
}

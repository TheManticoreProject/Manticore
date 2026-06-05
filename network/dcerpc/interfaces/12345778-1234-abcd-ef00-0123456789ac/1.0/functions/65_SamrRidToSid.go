package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// samrRidToSidRequest carries the [in] parameters of SamrRidToSid: an object
// handle for the domain context and the relative identifier to expand.
type samrRidToSidRequest struct {
	ObjectHandle structures.SAMPR_HANDLE
	Rid          ndr.DWORD
}

func (*samrRidToSidRequest) Opnum() uint16 { return samr.OpnumSamrRidToSid }

// samrRidToSidResponse carries the [out] double pointer to the constructed SID and
// the NTSTATUS.
type samrRidToSidResponse struct {
	Sid    *dtyp.RPC_SID `ndr:"unique"`
	Status ndr.DWORD     `ndr:"retval"`
}

// SamrRidToSid calls SamrRidToSid (opnum 65), constructing the full SID for a RID
// in the object's domain ([MS-SAMR] 3.1.5.11.2).
func SamrRidToSid(rpc ndr.Invoker, handle structures.SAMPR_HANDLE, rid uint32) (*dtyp.RPC_SID, error) {
	req := &samrRidToSidRequest{
		ObjectHandle: handle,
		Rid:          ndr.DWORD(rid),
	}
	var resp samrRidToSidResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("SamrRidToSid: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Sid, fmt.Errorf("SamrRidToSid failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Sid, nil
}

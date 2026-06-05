package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// lsarLookupPrivilegeNameRequest is the [in] parameter set of LsarLookupPrivilegeName:
// the policy handle and the privilege LUID value (a top-level [ref] pointer, so its value
// is inline).
type lsarLookupPrivilegeNameRequest struct {
	PolicyHandle structures.LSAPR_HANDLE
	Value        dtyp.LUID
}

func (*lsarLookupPrivilegeNameRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarLookupPrivilegeName
}

// lsarLookupPrivilegeNameResponse is the reply: the [out] privilege name (a [unique]
// pointer to an RPC_UNICODE_STRING, returned through a double pointer) and the NTSTATUS
// return value.
type lsarLookupPrivilegeNameResponse struct {
	Name   *dtyp.RPC_UNICODE_STRING `ndr:"unique"`
	Status ndr.DWORD                `ndr:"retval"`
}

// LsarLookupPrivilegeName calls LsarLookupPrivilegeName (opnum 32), mapping a privilege
// LUID to its name ([MS-LSAD] 3.1.4.5.12).
func LsarLookupPrivilegeName(rpc ndr.Invoker, policyHandle structures.LSAPR_HANDLE, value dtyp.LUID) (*dtyp.RPC_UNICODE_STRING, error) {
	req := &lsarLookupPrivilegeNameRequest{
		PolicyHandle: policyHandle,
		Value:        value,
	}
	var resp lsarLookupPrivilegeNameResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("LsarLookupPrivilegeName: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Name, fmt.Errorf("LsarLookupPrivilegeName failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Name, nil
}

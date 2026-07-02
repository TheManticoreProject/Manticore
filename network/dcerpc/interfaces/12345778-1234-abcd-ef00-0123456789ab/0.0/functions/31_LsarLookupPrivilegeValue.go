package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarLookupPrivilegeValueRequest is the [in] parameter set of LsarLookupPrivilegeValue:
// the policy handle and the privilege name. Name is a top-level [in] PRPC_UNICODE_STRING,
// i.e. a [ref] pointer ([C706]: a top-level parameter pointer with no explicit attribute
// is a reference pointer), so its referent is transmitted in place with no referent id —
// modeled as an inline value, not a [unique] pointer.
type lsarLookupPrivilegeValueRequest struct {
	PolicyHandle mslsad.LSAPR_HANDLE
	Name         dtyp.RPC_UNICODE_STRING
}

func (*lsarLookupPrivilegeValueRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarLookupPrivilegeValue
}

// lsarLookupPrivilegeValueResponse is the reply: the [out] LUID value of the privilege
// (a top-level [ref] pointer, so its value is inline) and the NTSTATUS return value.
type lsarLookupPrivilegeValueResponse struct {
	Value  dtyp.LUID
	Status ndr.DWORD `ndr:"retval"`
}

// LsarLookupPrivilegeValue calls LsarLookupPrivilegeValue (opnum 31), mapping a privilege
// name to its locally unique identifier ([MS-LSAD] 3.1.4.5.13).
func LsarLookupPrivilegeValue(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, name string) (dtyp.LUID, error) {
	req := &lsarLookupPrivilegeValueRequest{
		PolicyHandle: policyHandle,
		Name:         dtyp.NewUnicodeString(name),
	}
	var resp lsarLookupPrivilegeValueResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return dtyp.LUID{}, fmt.Errorf("LsarLookupPrivilegeValue: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Value, fmt.Errorf("LsarLookupPrivilegeValue failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Value, nil
}

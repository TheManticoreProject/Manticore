package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarOpenPolicyRequest is the [in] parameter set of LsarOpenPolicy (opnum 6): a NULL
// [unique] SystemName pointer, an inline LSAPR_OBJECT_ATTRIBUTES, and the desired access
// mask. It differs from LsarOpenPolicy2 only in that SystemName is a plain wchar_t*
// rather than a [string] wchar_t*; the server ignores it, so a NULL referent (nil
// pointer) marshals identically.
type lsarOpenPolicyRequest struct {
	SystemName    *ndr.WSTR `ndr:"unique"`
	Attributes    structures.LSAPR_OBJECT_ATTRIBUTES
	DesiredAccess ndr.DWORD
}

func (*lsarOpenPolicyRequest) Opnum() uint16 { return lsarpc.OpnumLsarOpenPolicy }

// LsarOpenPolicy calls LsarOpenPolicy (opnum 6) and returns a policy handle. Prefer
// LsarOpenPolicy2 (opnum 44) against modern servers; this is the legacy form.
func LsarOpenPolicy(rpc *client.Client, desiredAccess uint32) (structures.LSAPR_HANDLE, error) {
	req := &lsarOpenPolicyRequest{DesiredAccess: ndr.DWORD(desiredAccess)}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.LSAPR_HANDLE{}, fmt.Errorf("LsarOpenPolicy: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarOpenPolicy failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}

package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarOpenPolicy2Request is the [in] parameter set of LsarOpenPolicy2: a NULL unique
// SystemName pointer, an inline LSAPR_OBJECT_ATTRIBUTES (a top-level [ref] struct), and
// the desired access mask.
type lsarOpenPolicy2Request struct {
	SystemName    *ndr.WSTR `ndr:"unique"`
	Attributes    structures.LSAPR_OBJECT_ATTRIBUTES
	DesiredAccess ndr.DWORD
}

func (*lsarOpenPolicy2Request) Opnum() uint16 { return lsarpc.OpnumLsarOpenPolicy2 }

// LsarOpenPolicy2 calls LsarOpenPolicy2 (opnum 44) and returns a policy handle.
func LsarOpenPolicy2(rpc *client.Client, desiredAccess uint32) (structures.LSAPR_HANDLE, error) {
	req := &lsarOpenPolicy2Request{DesiredAccess: ndr.DWORD(desiredAccess)}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.LSAPR_HANDLE{}, fmt.Errorf("LsarOpenPolicy2: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarOpenPolicy2 failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}

package functions

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// samrGetMembersInAliasRequest carries the [in] alias handle.
type samrGetMembersInAliasRequest struct {
	AliasHandle structures.SAMPR_HANDLE
}

func (*samrGetMembersInAliasRequest) Opnum() uint16 { return samr.OpnumSamrGetMembersInAlias }

// samrGetMembersInAliasResponse is the reply: the [out] SID array container (a single [ref]
// out parameter, transmitted inline) and the NTSTATUS.
type samrGetMembersInAliasResponse struct {
	Members structures.SAMPR_PSID_ARRAY_OUT
	Status  ndr.DWORD `ndr:"retval"`
}

// SamrGetMembersInAlias calls SamrGetMembersInAlias (opnum 33), retrieving the SIDs of the
// members of an alias ([MS-SAMR] 3.1.5.5.5).
func SamrGetMembersInAlias(rpc ndr.Invoker, aliasHandle structures.SAMPR_HANDLE) (structures.SAMPR_PSID_ARRAY_OUT, error) {
	req := &samrGetMembersInAliasRequest{AliasHandle: aliasHandle}
	var resp samrGetMembersInAliasResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.SAMPR_PSID_ARRAY_OUT{}, fmt.Errorf("SamrGetMembersInAlias: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Members, fmt.Errorf("SamrGetMembersInAlias failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Members, nil
}

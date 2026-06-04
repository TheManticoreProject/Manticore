package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrAddMultipleMembersToAliasRequest carries the [in] alias handle and the [ref] SID
// array container of members to add (transmitted inline).
type samrAddMultipleMembersToAliasRequest struct {
	AliasHandle   structures.SAMPR_HANDLE
	MembersBuffer structures.SAMPR_PSID_ARRAY
}

func (*samrAddMultipleMembersToAliasRequest) Opnum() uint16 {
	return samr.OpnumSamrAddMultipleMembersToAlias
}

// SamrAddMultipleMembersToAlias calls SamrAddMultipleMembersToAlias (opnum 52), adding
// multiple members to an alias in a single call ([MS-SAMR] 3.1.5.4.3).
func SamrAddMultipleMembersToAlias(rpc *client.Client, aliasHandle structures.SAMPR_HANDLE, membersBuffer structures.SAMPR_PSID_ARRAY) error {
	req := &samrAddMultipleMembersToAliasRequest{
		AliasHandle:   aliasHandle,
		MembersBuffer: membersBuffer,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrAddMultipleMembersToAlias: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrAddMultipleMembersToAlias failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}

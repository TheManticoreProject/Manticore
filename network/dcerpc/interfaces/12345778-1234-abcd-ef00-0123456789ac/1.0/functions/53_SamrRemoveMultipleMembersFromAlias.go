package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrRemoveMultipleMembersFromAliasRequest carries the [in] alias handle and the [ref] SID
// array container of members to remove (transmitted inline).
type samrRemoveMultipleMembersFromAliasRequest struct {
	AliasHandle   structures.SAMPR_HANDLE
	MembersBuffer structures.SAMPR_PSID_ARRAY
}

func (*samrRemoveMultipleMembersFromAliasRequest) Opnum() uint16 {
	return samr.OpnumSamrRemoveMultipleMembersFromAlias
}

// SamrRemoveMultipleMembersFromAlias calls SamrRemoveMultipleMembersFromAlias (opnum 53),
// removing multiple members from an alias in a single call ([MS-SAMR] 3.1.5.5.4).
func SamrRemoveMultipleMembersFromAlias(rpc ndr.Invoker, aliasHandle structures.SAMPR_HANDLE, membersBuffer structures.SAMPR_PSID_ARRAY) error {
	req := &samrRemoveMultipleMembersFromAliasRequest{
		AliasHandle:   aliasHandle,
		MembersBuffer: membersBuffer,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrRemoveMultipleMembersFromAlias: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrRemoveMultipleMembersFromAlias failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}

package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrDeleteAliasRequest carries the [in,out] SAMPR_HANDLE of the alias to delete. On
// success the server returns it zeroed via the shared handleResponse.
type samrDeleteAliasRequest struct {
	AliasHandle structures.SAMPR_HANDLE
}

func (*samrDeleteAliasRequest) Opnum() uint16 { return samr.OpnumSamrDeleteAlias }

// SamrDeleteAlias calls SamrDeleteAlias (opnum 30), removing an alias from the database and
// returning the (now zeroed) handle ([MS-SAMR] 3.1.5.7.2).
func SamrDeleteAlias(rpc *client.Client, aliasHandle structures.SAMPR_HANDLE) (structures.SAMPR_HANDLE, error) {
	req := &samrDeleteAliasRequest{AliasHandle: aliasHandle}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return aliasHandle, fmt.Errorf("SamrDeleteAlias: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Handle, fmt.Errorf("SamrDeleteAlias failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}

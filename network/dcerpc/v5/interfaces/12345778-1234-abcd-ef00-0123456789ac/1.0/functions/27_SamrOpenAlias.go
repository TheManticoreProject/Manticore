package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrOpenAliasRequest carries the [in] domain handle, the desired access mask, and the
// relative id of the alias to open.
type samrOpenAliasRequest struct {
	DomainHandle  structures.SAMPR_HANDLE
	DesiredAccess ndr.DWORD
	AliasId       ndr.DWORD
}

func (*samrOpenAliasRequest) Opnum() uint16 { return samr.OpnumSamrOpenAlias }

// SamrOpenAlias calls SamrOpenAlias (opnum 27), obtaining a handle to an existing alias
// ([MS-SAMR] 3.1.5.1.6).
func SamrOpenAlias(rpc *client.Client, domainHandle structures.SAMPR_HANDLE, desiredAccess uint32, aliasId uint32) (structures.SAMPR_HANDLE, error) {
	req := &samrOpenAliasRequest{
		DomainHandle:  domainHandle,
		DesiredAccess: ndr.DWORD(desiredAccess),
		AliasId:       ndr.DWORD(aliasId),
	}
	var resp openHandleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.SAMPR_HANDLE{}, fmt.Errorf("SamrOpenAlias: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Handle, fmt.Errorf("SamrOpenAlias failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}

package functions

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrGetAliasMembershipRequest is the [in] parameter set of SamrGetAliasMembership: a
// domain handle and the [ref] SID array (single pointer container, inline) to test.
type samrGetAliasMembershipRequest struct {
	DomainHandle mssamr.SAMPR_HANDLE
	SidArray     mssamr.SAMPR_PSID_ARRAY
}

func (*samrGetAliasMembershipRequest) Opnum() uint16 { return samr.OpnumSamrGetAliasMembership }

// samrGetAliasMembershipResponse is the reply: the [out] [ref] SAMPR_ULONG_ARRAY of RIDs
// (single pointer container, inline) and the NTSTATUS.
type samrGetAliasMembershipResponse struct {
	Membership mssamr.SAMPR_ULONG_ARRAY
	Status     ndr.DWORD `ndr:"retval"`
}

// SamrGetAliasMembership calls SamrGetAliasMembership (opnum 16), returning the RIDs of the
// aliases in the domain of which any of the given SIDs is a member ([MS-SAMR] 3.1.5.9.2).
func SamrGetAliasMembership(rpc ndr.Invoker, domainHandle mssamr.SAMPR_HANDLE, sidArray mssamr.SAMPR_PSID_ARRAY) (mssamr.SAMPR_ULONG_ARRAY, error) {
	req := &samrGetAliasMembershipRequest{
		DomainHandle: domainHandle,
		SidArray:     sidArray,
	}
	var resp samrGetAliasMembershipResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssamr.SAMPR_ULONG_ARRAY{}, fmt.Errorf("SamrGetAliasMembership: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Membership, fmt.Errorf("SamrGetAliasMembership failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Membership, nil
}

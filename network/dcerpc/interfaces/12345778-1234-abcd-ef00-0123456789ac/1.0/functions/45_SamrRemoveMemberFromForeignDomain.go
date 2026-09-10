package functions

// IDL source: [MS-SAMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/1cd138b9-cc1b-4706-b115-49e53189e32e
// A fetched copy is kept at ms-samr.idl in the interface directory.

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrRemoveMemberFromForeignDomainRequest is the [in] parameter set of
// SamrRemoveMemberFromForeignDomain: a domain handle and the [ref] SID (inline, single
// pointer) to remove from all aliases in the domain.
type samrRemoveMemberFromForeignDomainRequest struct {
	DomainHandle mssamr.SAMPR_HANDLE
	MemberSid    msdtyp.RPC_SID
}

func (*samrRemoveMemberFromForeignDomainRequest) Opnum() uint16 {
	return samr.OpnumSamrRemoveMemberFromForeignDomain
}

// SamrRemoveMemberFromForeignDomain calls SamrRemoveMemberFromForeignDomain (opnum 45),
// removing the given SID from the membership of all aliases in the domain ([MS-SAMR]
// 3.1.5.8.5).
func SamrRemoveMemberFromForeignDomain(rpc ndr.Invoker, domainHandle mssamr.SAMPR_HANDLE, memberSid msdtyp.RPC_SID) error {
	req := &samrRemoveMemberFromForeignDomainRequest{
		DomainHandle: domainHandle,
		MemberSid:    memberSid,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrRemoveMemberFromForeignDomain: %w", err)
	}
	if nt_status.NT_STATUS(resp.Status) != nt_status.NT_STATUS_SUCCESS {
		return fmt.Errorf("SamrRemoveMemberFromForeignDomain failed: %s", nt_status.NT_STATUS(resp.Status).String())
	}
	return nil
}

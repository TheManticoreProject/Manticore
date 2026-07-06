package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarEnumerateAccountsRequest is the [in]/[in,out] parameter set of LsarEnumerateAccounts:
// an open policy handle, the [in,out] enumeration context (a cursor that the server
// updates between calls), and the maximum number of bytes the client will accept.
type lsarEnumerateAccountsRequest struct {
	PolicyHandle          mslsad.LSAPR_HANDLE
	EnumerationContext    ndr.DWORD
	PreferedMaximumLength ndr.DWORD
}

func (*lsarEnumerateAccountsRequest) Opnum() uint16 { return lsarpc.OpnumLsarEnumerateAccounts }

// lsarEnumerateAccountsResponse is the reply: the updated [in,out] enumeration context, the
// [out] buffer of account SIDs, and the NTSTATUS return value.
type lsarEnumerateAccountsResponse struct {
	EnumerationContext ndr.DWORD
	EnumerationBuffer  mslsad.LSAPR_ACCOUNT_ENUM_BUFFER
	Status             ndr.DWORD `ndr:"retval"`
}

// LsarEnumerateAccounts calls LsarEnumerateAccounts (opnum 11), returning a page of account
// SIDs known to the policy ([MS-LSAD] 3.1.4.5.2). The enumeration is stateful: pass the
// returned context back on the next call to continue, starting from 0. The server returns
// STATUS_MORE_ENTRIES while pages remain, which is not treated as an error; the returned
// context and buffer are valid in that case. The updated context is returned alongside the
// buffer.
func LsarEnumerateAccounts(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, enumerationContext uint32, preferedMaximumLength uint32) (mslsad.LSAPR_ACCOUNT_ENUM_BUFFER, uint32, error) {
	req := &lsarEnumerateAccountsRequest{
		PolicyHandle:          policyHandle,
		EnumerationContext:    ndr.DWORD(enumerationContext),
		PreferedMaximumLength: ndr.DWORD(preferedMaximumLength),
	}
	var resp lsarEnumerateAccountsResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mslsad.LSAPR_ACCOUNT_ENUM_BUFFER{}, enumerationContext, fmt.Errorf("LsarEnumerateAccounts: %w", err)
	}
	status := uint32(resp.Status)
	if status != lsarpc.StatusSuccess && status != lsarpc.StatusMoreEntries {
		return resp.EnumerationBuffer, uint32(resp.EnumerationContext), fmt.Errorf("LsarEnumerateAccounts failed: %s", lsarpc.StatusString(status))
	}
	return resp.EnumerationBuffer, uint32(resp.EnumerationContext), nil
}

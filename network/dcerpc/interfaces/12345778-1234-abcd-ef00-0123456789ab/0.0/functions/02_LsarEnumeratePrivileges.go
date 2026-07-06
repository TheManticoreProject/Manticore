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

// lsarEnumeratePrivilegesRequest is the [in]/[in,out] parameter set of
// LsarEnumeratePrivileges: the policy handle, the [in,out] enumeration context, and the
// preferred maximum length of the returned buffer.
type lsarEnumeratePrivilegesRequest struct {
	PolicyHandle          mslsad.LSAPR_HANDLE
	EnumerationContext    ndr.DWORD
	PreferedMaximumLength ndr.DWORD
}

func (*lsarEnumeratePrivilegesRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarEnumeratePrivileges
}

// lsarEnumeratePrivilegesResponse is the reply: the updated [in,out] enumeration context,
// the [out] enumeration buffer (a top-level [ref] pointer, so its value is inline), and
// the NTSTATUS return value.
type lsarEnumeratePrivilegesResponse struct {
	EnumerationContext ndr.DWORD
	EnumerationBuffer  mslsad.LSAPR_PRIVILEGE_ENUM_BUFFER
	Status             ndr.DWORD `ndr:"retval"`
}

// LsarEnumeratePrivileges calls LsarEnumeratePrivileges (opnum 2), enumerating the
// privileges known to the server ([MS-LSAD] 3.1.4.5.10). The caller passes the current
// enumeration context (0 to start) and the preferred maximum buffer length; it receives
// the updated context and the buffer of privilege definitions.
func LsarEnumeratePrivileges(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, enumerationContext uint32, preferedMaximumLength uint32) (uint32, mslsad.LSAPR_PRIVILEGE_ENUM_BUFFER, error) {
	req := &lsarEnumeratePrivilegesRequest{
		PolicyHandle:          policyHandle,
		EnumerationContext:    ndr.DWORD(enumerationContext),
		PreferedMaximumLength: ndr.DWORD(preferedMaximumLength),
	}
	var resp lsarEnumeratePrivilegesResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return 0, mslsad.LSAPR_PRIVILEGE_ENUM_BUFFER{}, fmt.Errorf("LsarEnumeratePrivileges: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return uint32(resp.EnumerationContext), resp.EnumerationBuffer, fmt.Errorf("LsarEnumeratePrivileges failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return uint32(resp.EnumerationContext), resp.EnumerationBuffer, nil
}

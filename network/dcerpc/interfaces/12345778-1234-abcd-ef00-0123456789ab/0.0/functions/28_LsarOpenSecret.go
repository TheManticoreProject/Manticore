package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarOpenSecretRequest is the [in] parameter set of LsarOpenSecret: an open policy
// handle, the secret name (a [ref] PRPC_UNICODE_STRING, modeled inline), and the desired
// access mask.
type lsarOpenSecretRequest struct {
	PolicyHandle  mslsad.LSAPR_HANDLE
	SecretName    msdtyp.RPC_UNICODE_STRING
	DesiredAccess ndr.DWORD
}

func (*lsarOpenSecretRequest) Opnum() uint16 { return lsarpc.OpnumLsarOpenSecret }

// LsarOpenSecret calls LsarOpenSecret (opnum 28) and returns a handle to the named secret
// object ([MS-LSAD] 3.1.4.6.2).
func LsarOpenSecret(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, secretName string, desiredAccess uint32) (mslsad.LSAPR_HANDLE, error) {
	req := &lsarOpenSecretRequest{
		PolicyHandle:  policyHandle,
		SecretName:    msdtyp.NewUnicodeString(secretName),
		DesiredAccess: ndr.DWORD(desiredAccess),
	}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mslsad.LSAPR_HANDLE{}, fmt.Errorf("LsarOpenSecret: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarOpenSecret failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}

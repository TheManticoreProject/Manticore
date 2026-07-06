package functions

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarLookupPrivilegeDisplayNameRequest is the [in] parameter set of
// LsarLookupPrivilegeDisplayName: the policy handle, the privilege name (a [unique]
// pointer to an RPC_UNICODE_STRING), and the client's language identifiers.
type lsarLookupPrivilegeDisplayNameRequest struct {
	PolicyHandle                mslsad.LSAPR_HANDLE
	Name                        msdtyp.RPC_UNICODE_STRING
	ClientLanguage              int16
	ClientSystemDefaultLanguage int16
}

func (*lsarLookupPrivilegeDisplayNameRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarLookupPrivilegeDisplayName
}

// lsarLookupPrivilegeDisplayNameResponse is the reply: the [out] display name (a [unique]
// pointer to an RPC_UNICODE_STRING, returned through a double pointer), the [out] language
// actually returned (a top-level [ref] pointer, so its value is inline), and the NTSTATUS
// return value.
type lsarLookupPrivilegeDisplayNameResponse struct {
	DisplayName      *msdtyp.RPC_UNICODE_STRING `ndr:"unique"`
	LanguageReturned uint16
	Status           ndr.DWORD `ndr:"retval"`
}

// LsarLookupPrivilegeDisplayName calls LsarLookupPrivilegeDisplayName (opnum 33), mapping a
// privilege name to its localized display name ([MS-LSAD] 3.1.4.5.11). It returns the
// display name and the language identifier the server actually used.
func LsarLookupPrivilegeDisplayName(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, name string, clientLanguage int16, clientSystemDefaultLanguage int16) (*msdtyp.RPC_UNICODE_STRING, uint16, error) {
	rpcName := msdtyp.NewUnicodeString(name)
	req := &lsarLookupPrivilegeDisplayNameRequest{
		PolicyHandle:                policyHandle,
		Name:                        rpcName,
		ClientLanguage:              clientLanguage,
		ClientSystemDefaultLanguage: clientSystemDefaultLanguage,
	}
	var resp lsarLookupPrivilegeDisplayNameResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, 0, fmt.Errorf("LsarLookupPrivilegeDisplayName: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.DisplayName, resp.LanguageReturned, fmt.Errorf("LsarLookupPrivilegeDisplayName failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.DisplayName, resp.LanguageReturned, nil
}

package functions

// IDL source: [MS-NRPC] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7
// A fetched copy is kept at ms-nrpc.idl in the interface directory.

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// netrDatabaseSyncRequest carries the [in] parameters of NetrDatabaseSync.
type netrDatabaseSyncRequest struct {
	PrimaryName            ndr.WSTR
	ComputerName           ndr.WSTR
	Authenticator          msnrpc.NETLOGON_AUTHENTICATOR
	ReturnAuthenticator    msnrpc.NETLOGON_AUTHENTICATOR
	DatabaseID             ndr.DWORD
	SyncContext            ndr.DWORD
	PreferredMaximumLength ndr.DWORD
}

func (*netrDatabaseSyncRequest) Opnum() uint16 { return logon.OpnumNetrDatabaseSync }

// netrDatabaseSyncResponse carries the [out] parameters and return value of NetrDatabaseSync.
type netrDatabaseSyncResponse struct {
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	SyncContext         ndr.DWORD
	DeltaArray          *msnrpc.NETLOGON_DELTA_ENUM_ARRAY `ndr:"unique"`
	Status              ndr.DWORD                         `ndr:"retval"`
}

// NetrDatabaseSync calls NetrDatabaseSync (opnum 8) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrDatabaseSync(rpc ndr.Invoker, primaryName ndr.WSTR, computerName ndr.WSTR, authenticator msnrpc.NETLOGON_AUTHENTICATOR, returnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, databaseID ndr.DWORD, syncContext ndr.DWORD, preferredMaximumLength ndr.DWORD) (ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, SyncContext ndr.DWORD, DeltaArray *msnrpc.NETLOGON_DELTA_ENUM_ARRAY, err error) {
	req := &netrDatabaseSyncRequest{
		PrimaryName:            primaryName,
		ComputerName:           computerName,
		Authenticator:          authenticator,
		ReturnAuthenticator:    returnAuthenticator,
		DatabaseID:             databaseID,
		SyncContext:            syncContext,
		PreferredMaximumLength: preferredMaximumLength,
	}
	var resp netrDatabaseSyncResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDatabaseSync: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	SyncContext = resp.SyncContext
	DeltaArray = resp.DeltaArray
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrDatabaseSync failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

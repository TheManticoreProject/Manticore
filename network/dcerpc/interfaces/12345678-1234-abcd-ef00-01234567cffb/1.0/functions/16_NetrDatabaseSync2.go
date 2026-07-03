package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// netrDatabaseSync2Request carries the [in] parameters of NetrDatabaseSync2.
type netrDatabaseSync2Request struct {
	PrimaryName            ndr.WSTR
	ComputerName           ndr.WSTR
	Authenticator          msnrpc.NETLOGON_AUTHENTICATOR
	ReturnAuthenticator    msnrpc.NETLOGON_AUTHENTICATOR
	DatabaseID             ndr.DWORD
	RestartState           msnrpc.SYNC_STATE
	SyncContext            ndr.DWORD
	PreferredMaximumLength ndr.DWORD
}

func (*netrDatabaseSync2Request) Opnum() uint16 { return logon.OpnumNetrDatabaseSync2 }

// netrDatabaseSync2Response carries the [out] parameters and return value of NetrDatabaseSync2.
type netrDatabaseSync2Response struct {
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	SyncContext         ndr.DWORD
	DeltaArray          *msnrpc.NETLOGON_DELTA_ENUM_ARRAY `ndr:"unique"`
	Status              ndr.DWORD                         `ndr:"retval"`
}

// NetrDatabaseSync2 calls NetrDatabaseSync2 (opnum 16) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrDatabaseSync2(rpc ndr.Invoker, primaryName ndr.WSTR, computerName ndr.WSTR, authenticator msnrpc.NETLOGON_AUTHENTICATOR, returnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, databaseID ndr.DWORD, restartState msnrpc.SYNC_STATE, syncContext ndr.DWORD, preferredMaximumLength ndr.DWORD) (ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, SyncContext ndr.DWORD, DeltaArray *msnrpc.NETLOGON_DELTA_ENUM_ARRAY, err error) {
	req := &netrDatabaseSync2Request{
		PrimaryName:            primaryName,
		ComputerName:           computerName,
		Authenticator:          authenticator,
		ReturnAuthenticator:    returnAuthenticator,
		DatabaseID:             databaseID,
		RestartState:           restartState,
		SyncContext:            syncContext,
		PreferredMaximumLength: preferredMaximumLength,
	}
	var resp netrDatabaseSync2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDatabaseSync2: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	SyncContext = resp.SyncContext
	DeltaArray = resp.DeltaArray
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrDatabaseSync2 failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

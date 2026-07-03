package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// netrDatabaseDeltasRequest carries the [in] parameters of NetrDatabaseDeltas.
type netrDatabaseDeltasRequest struct {
	PrimaryName            ndr.WSTR
	ComputerName           ndr.WSTR
	Authenticator          msnrpc.NETLOGON_AUTHENTICATOR
	ReturnAuthenticator    msnrpc.NETLOGON_AUTHENTICATOR
	DatabaseID             ndr.DWORD
	DomainModifiedCount    msnrpc.NLPR_MODIFIED_COUNT
	PreferredMaximumLength ndr.DWORD
}

func (*netrDatabaseDeltasRequest) Opnum() uint16 { return logon.OpnumNetrDatabaseDeltas }

// netrDatabaseDeltasResponse carries the [out] parameters and return value of NetrDatabaseDeltas.
type netrDatabaseDeltasResponse struct {
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	DomainModifiedCount msnrpc.NLPR_MODIFIED_COUNT
	DeltaArray          *msnrpc.NETLOGON_DELTA_ENUM_ARRAY `ndr:"unique"`
	Status              ndr.DWORD                         `ndr:"retval"`
}

// NetrDatabaseDeltas calls NetrDatabaseDeltas (opnum 7) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrDatabaseDeltas(rpc ndr.Invoker, primaryName ndr.WSTR, computerName ndr.WSTR, authenticator msnrpc.NETLOGON_AUTHENTICATOR, returnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, databaseID ndr.DWORD, domainModifiedCount msnrpc.NLPR_MODIFIED_COUNT, preferredMaximumLength ndr.DWORD) (ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, DomainModifiedCount msnrpc.NLPR_MODIFIED_COUNT, DeltaArray *msnrpc.NETLOGON_DELTA_ENUM_ARRAY, err error) {
	req := &netrDatabaseDeltasRequest{
		PrimaryName:            primaryName,
		ComputerName:           computerName,
		Authenticator:          authenticator,
		ReturnAuthenticator:    returnAuthenticator,
		DatabaseID:             databaseID,
		DomainModifiedCount:    domainModifiedCount,
		PreferredMaximumLength: preferredMaximumLength,
	}
	var resp netrDatabaseDeltasResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDatabaseDeltas: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	DomainModifiedCount = resp.DomainModifiedCount
	DeltaArray = resp.DeltaArray
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrDatabaseDeltas failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

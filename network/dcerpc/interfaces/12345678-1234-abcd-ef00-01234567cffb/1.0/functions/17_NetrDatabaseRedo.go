package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// netrDatabaseRedoRequest carries the [in] parameters of NetrDatabaseRedo.
type netrDatabaseRedoRequest struct {
	PrimaryName         ndr.WSTR
	ComputerName        ndr.WSTR
	Authenticator       msnrpc.NETLOGON_AUTHENTICATOR
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	ChangeLogEntry      []uint8 `ndr:"ref,size_is=ChangeLogEntrySize"`
	ChangeLogEntrySize  ndr.DWORD
}

func (*netrDatabaseRedoRequest) Opnum() uint16 { return logon.OpnumNetrDatabaseRedo }

// netrDatabaseRedoResponse carries the [out] parameters and return value of NetrDatabaseRedo.
type netrDatabaseRedoResponse struct {
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	DeltaArray          *msnrpc.NETLOGON_DELTA_ENUM_ARRAY `ndr:"unique"`
	Status              ndr.DWORD                         `ndr:"retval"`
}

// NetrDatabaseRedo calls NetrDatabaseRedo (opnum 17) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrDatabaseRedo(rpc ndr.Invoker, primaryName ndr.WSTR, computerName ndr.WSTR, authenticator msnrpc.NETLOGON_AUTHENTICATOR, returnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, changeLogEntry []uint8, changeLogEntrySize ndr.DWORD) (ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, DeltaArray *msnrpc.NETLOGON_DELTA_ENUM_ARRAY, err error) {
	req := &netrDatabaseRedoRequest{
		PrimaryName:         primaryName,
		ComputerName:        computerName,
		Authenticator:       authenticator,
		ReturnAuthenticator: returnAuthenticator,
		ChangeLogEntry:      changeLogEntry,
		ChangeLogEntrySize:  changeLogEntrySize,
	}
	var resp netrDatabaseRedoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDatabaseRedo: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	DeltaArray = resp.DeltaArray
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrDatabaseRedo failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

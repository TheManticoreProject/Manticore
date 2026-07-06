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

// netrAccountSyncRequest carries the [in] parameters of NetrAccountSync.
type netrAccountSyncRequest struct {
	PrimaryName         *ndr.WSTR `ndr:"unique"`
	ComputerName        ndr.WSTR
	Authenticator       msnrpc.NETLOGON_AUTHENTICATOR
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	Reference           ndr.DWORD
	Level               ndr.DWORD
	BufferSize          ndr.DWORD
}

func (*netrAccountSyncRequest) Opnum() uint16 { return logon.OpnumNetrAccountSync }

// netrAccountSyncResponse carries the [out] parameters and return value of NetrAccountSync.
type netrAccountSyncResponse struct {
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	Buffer              []uint8 `ndr:"ref,size_is=BufferSize"`
	CountReturned       ndr.DWORD
	TotalEntries        ndr.DWORD
	NextReference       ndr.DWORD
	LastRecordId        msnrpc.UAS_INFO_0
	Status              ndr.DWORD `ndr:"retval"`
}

// NetrAccountSync calls NetrAccountSync (opnum 10) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrAccountSync(rpc ndr.Invoker, primaryName *ndr.WSTR, computerName ndr.WSTR, authenticator msnrpc.NETLOGON_AUTHENTICATOR, returnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, reference ndr.DWORD, level ndr.DWORD, bufferSize ndr.DWORD) (ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, Buffer []uint8, CountReturned ndr.DWORD, TotalEntries ndr.DWORD, NextReference ndr.DWORD, LastRecordId msnrpc.UAS_INFO_0, err error) {
	req := &netrAccountSyncRequest{
		PrimaryName:         primaryName,
		ComputerName:        computerName,
		Authenticator:       authenticator,
		ReturnAuthenticator: returnAuthenticator,
		Reference:           reference,
		Level:               level,
		BufferSize:          bufferSize,
	}
	var resp netrAccountSyncResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrAccountSync: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	Buffer = resp.Buffer
	CountReturned = resp.CountReturned
	TotalEntries = resp.TotalEntries
	NextReference = resp.NextReference
	LastRecordId = resp.LastRecordId
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrAccountSync failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

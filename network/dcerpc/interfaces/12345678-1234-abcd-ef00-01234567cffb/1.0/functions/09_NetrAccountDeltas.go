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

// netrAccountDeltasRequest carries the [in] parameters of NetrAccountDeltas.
type netrAccountDeltasRequest struct {
	PrimaryName         *ndr.WSTR `ndr:"unique"`
	ComputerName        ndr.WSTR
	Authenticator       msnrpc.NETLOGON_AUTHENTICATOR
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	RecordId            msnrpc.UAS_INFO_0
	Count               ndr.DWORD
	Level               ndr.DWORD
	BufferSize          ndr.DWORD
}

func (*netrAccountDeltasRequest) Opnum() uint16 { return logon.OpnumNetrAccountDeltas }

// netrAccountDeltasResponse carries the [out] parameters and return value of NetrAccountDeltas.
type netrAccountDeltasResponse struct {
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	Buffer              []uint8 `ndr:"ref,size_is=BufferSize"`
	CountReturned       ndr.DWORD
	TotalEntries        ndr.DWORD
	NextRecordId        msnrpc.UAS_INFO_0
	Status              ndr.DWORD `ndr:"retval"`
}

// NetrAccountDeltas calls NetrAccountDeltas (opnum 9) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrAccountDeltas(rpc ndr.Invoker, primaryName *ndr.WSTR, computerName ndr.WSTR, authenticator msnrpc.NETLOGON_AUTHENTICATOR, returnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, recordId msnrpc.UAS_INFO_0, count ndr.DWORD, level ndr.DWORD, bufferSize ndr.DWORD) (ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, Buffer []uint8, CountReturned ndr.DWORD, TotalEntries ndr.DWORD, NextRecordId msnrpc.UAS_INFO_0, err error) {
	req := &netrAccountDeltasRequest{
		PrimaryName:         primaryName,
		ComputerName:        computerName,
		Authenticator:       authenticator,
		ReturnAuthenticator: returnAuthenticator,
		RecordId:            recordId,
		Count:               count,
		Level:               level,
		BufferSize:          bufferSize,
	}
	var resp netrAccountDeltasResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrAccountDeltas: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	Buffer = resp.Buffer
	CountReturned = resp.CountReturned
	TotalEntries = resp.TotalEntries
	NextRecordId = resp.NextRecordId
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrAccountDeltas failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

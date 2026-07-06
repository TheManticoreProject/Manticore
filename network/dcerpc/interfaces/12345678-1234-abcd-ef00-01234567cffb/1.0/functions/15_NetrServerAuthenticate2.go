package functions

// IDL source: [MS-NRPC] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7
// A fetched copy is kept at ms-nrpc.idl in the interface directory.

import (
	"fmt"

	netlogon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

type netrServerAuthenticate2Request struct {
	PrimaryName       *ndr.WSTR `ndr:"unique"`
	AccountName       ndr.WSTR
	SecureChannelType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE `ndr:"enum"`
	ComputerName      ndr.WSTR
	ClientCredential  msnrpc.NETLOGON_CREDENTIAL
	NegotiateFlags    ndr.DWORD
}

func (*netrServerAuthenticate2Request) Opnum() uint16 {
	return netlogon.OpnumNetrServerAuthenticate2
}

type netrServerAuthenticate2Response struct {
	ServerCredential msnrpc.NETLOGON_CREDENTIAL
	NegotiateFlags   ndr.DWORD
	Status           ndr.DWORD `ndr:"retval"`
}

// NetrServerAuthenticate2 calls NetrServerAuthenticate2 ([MS-NRPC] 3.5.4.4.2, opnum 15). It
// returns the server-computed credential, the negotiated flags, and the raw NTSTATUS, so
// the caller can distinguish STATUS_ACCESS_DENIED from STATUS_SUCCESS without treating the
// former as a transport error.
//
// Parameters:
//   - rpc: A DCE/RPC client bound to the Netlogon interface.
//   - primaryName: The DC name (empty sends a NULL unique pointer).
//   - accountName: The account being authenticated (e.g. "DC$").
//   - secureChannelType: The secure channel kind.
//   - computerName: The NetBIOS name of the client computer.
//   - clientCredential: The client credential to present.
//   - negotiateFlags: The requested negotiate flags.
//
// Returns:
//   - The server credential, the negotiated flags, the NTSTATUS, and a transport error.
func NetrServerAuthenticate2(rpc ndr.Invoker, primaryName, accountName string, secureChannelType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE, computerName string, clientCredential msnrpc.NETLOGON_CREDENTIAL, negotiateFlags uint32) (msnrpc.NETLOGON_CREDENTIAL, uint32, uint32, error) {
	req := &netrServerAuthenticate2Request{
		AccountName:       ndr.WSTR(accountName),
		SecureChannelType: secureChannelType,
		ComputerName:      ndr.WSTR(computerName),
		ClientCredential:  clientCredential,
		NegotiateFlags:    ndr.DWORD(negotiateFlags),
	}
	if primaryName != "" {
		pn := ndr.WSTR(primaryName)
		req.PrimaryName = &pn
	}

	var resp netrServerAuthenticate2Response
	if err := rpc.Invoke(req, &resp); err != nil {
		return msnrpc.NETLOGON_CREDENTIAL{}, 0, 0, fmt.Errorf("NetrServerAuthenticate2: %w", err)
	}
	return resp.ServerCredential, uint32(resp.NegotiateFlags), uint32(resp.Status), nil
}

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

// netrChainSetClientAttributesRequest carries the [in] parameters of NetrChainSetClientAttributes.
type netrChainSetClientAttributesRequest struct {
	PrimaryName           ndr.WSTR
	ChainedFromServerName ndr.WSTR
	ChainedForClientName  ndr.WSTR
	Authenticator         msnrpc.NETLOGON_AUTHENTICATOR
	ReturnAuthenticator   msnrpc.NETLOGON_AUTHENTICATOR
	DwInVersion           ndr.DWORD
	PmsgIn                msnrpc.NL_IN_CHAIN_SET_CLIENT_ATTRIBUTES
	PdwOutVersion         ndr.DWORD
	PmsgOut               msnrpc.NL_OUT_CHAIN_SET_CLIENT_ATTRIBUTES
}

func (*netrChainSetClientAttributesRequest) Opnum() uint16 {
	return logon.OpnumNetrChainSetClientAttributes
}

// netrChainSetClientAttributesResponse carries the [out] parameters and return value of NetrChainSetClientAttributes.
type netrChainSetClientAttributesResponse struct {
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	PdwOutVersion       ndr.DWORD
	PmsgOut             msnrpc.NL_OUT_CHAIN_SET_CLIENT_ATTRIBUTES
	Status              ndr.DWORD `ndr:"retval"`
}

// NetrChainSetClientAttributes calls NetrChainSetClientAttributes (opnum 49) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrChainSetClientAttributes(rpc ndr.Invoker, primaryName ndr.WSTR, chainedFromServerName ndr.WSTR, chainedForClientName ndr.WSTR, authenticator msnrpc.NETLOGON_AUTHENTICATOR, returnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, dwInVersion ndr.DWORD, pmsgIn msnrpc.NL_IN_CHAIN_SET_CLIENT_ATTRIBUTES, pdwOutVersion ndr.DWORD, pmsgOut msnrpc.NL_OUT_CHAIN_SET_CLIENT_ATTRIBUTES) (ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, PdwOutVersion ndr.DWORD, PmsgOut msnrpc.NL_OUT_CHAIN_SET_CLIENT_ATTRIBUTES, err error) {
	req := &netrChainSetClientAttributesRequest{
		PrimaryName:           primaryName,
		ChainedFromServerName: chainedFromServerName,
		ChainedForClientName:  chainedForClientName,
		Authenticator:         authenticator,
		ReturnAuthenticator:   returnAuthenticator,
		DwInVersion:           dwInVersion,
		PmsgIn:                pmsgIn,
		PdwOutVersion:         pdwOutVersion,
		PmsgOut:               pmsgOut,
	}
	var resp netrChainSetClientAttributesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrChainSetClientAttributes: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrChainSetClientAttributes failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

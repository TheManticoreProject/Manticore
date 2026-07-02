package functions

import (
	"fmt"

	BitsPeerAuth "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3d0d746-d2af-40fd-8a7a-0d7078bb7092/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msbpau "github.com/TheManticoreProject/Manticore/windows/protocols/ms-bpau"
)

// exchangePublicKeysRequest carries the [in] parameters of ExchangePublicKeys.
//
// The IDL binding handle ([in] handle_t Binding) is the explicit RPC binding and is
// not marshalled, so it is omitted. ClientKey is a [unique] pointer to a conformant
// byte array whose maximum_count is ClientKeyLength; a nil ClientKey (ClientKeyLength
// zero) is transmitted as a null referent, meaning "not sending a certificate".
type exchangePublicKeysRequest struct {
	ClientKeyLength msbpau.KEY_LENGTH
	ClientKey       []byte `ndr:"unique,size_is=ClientKeyLength"`
}

func (*exchangePublicKeysRequest) Opnum() uint16 { return BitsPeerAuth.OpnumExchangePublicKeys }

// exchangePublicKeysResponse carries the [out] parameters and HRESULT return value.
//
// pServerKeyLength ([out, ref] KEY_LENGTH*) is a single ref pointer to a scalar, so it
// is transmitted inline. pServerKey ([out, ref] byte** with size_is(, *pServerKeyLength))
// is a ref pointer to a [unique] pointer to a conformant byte array: the outer ref adds
// nothing to the wire, and the inner unique pointer emits a referent id followed by the
// deferred conformant array of PServerKeyLength bytes. Status is the trailing HRESULT
// (encoded after the deferred referents, hence the retval tag).
type exchangePublicKeysResponse struct {
	PServerKeyLength msbpau.KEY_LENGTH
	PServerKey       []byte    `ndr:"unique,size_is=PServerKeyLength"`
	Status           ndr.DWORD `ndr:"retval"`
}

// ExchangePublicKeys calls ExchangePublicKeys (opnum 0) ([MS-BPAU] 3.2.4.1). The client
// offers its local certificate (a CERTIFICATE_BLOB, [MS-BPAU] 2.2.2) in clientKey — pass
// nil to decline — and receives the server's certificate blob in return. The method
// returns an HRESULT; ERROR_SUCCESS (0x00000000) is success, and the server returns
// 0x80070005 (E_ACCESSDENIED) when the caller's Kerberos identity is not trusted.
func ExchangePublicKeys(rpc ndr.Invoker, clientKey []byte) ([]byte, error) {
	req := &exchangePublicKeysRequest{
		ClientKeyLength: msbpau.KEY_LENGTH(len(clientKey)),
		ClientKey:       clientKey,
	}
	var resp exchangePublicKeysResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("ExchangePublicKeys: %w", err)
	}
	if uint32(resp.Status) != BitsPeerAuth.StatusSuccess {
		return nil, fmt.Errorf("ExchangePublicKeys failed: %s", BitsPeerAuth.StatusString(uint32(resp.Status)))
	}
	return resp.PServerKey, nil
}

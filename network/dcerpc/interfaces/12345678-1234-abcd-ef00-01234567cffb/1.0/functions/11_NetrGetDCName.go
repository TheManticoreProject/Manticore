package functions

// IDL source: [MS-NRPC] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7
// A fetched copy is kept at ms-nrpc.idl in the interface directory.

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrGetDCNameRequest carries the [in] parameters of NetrGetDCName.
type netrGetDCNameRequest struct {
	ServerName ndr.WSTR
	DomainName *ndr.WSTR `ndr:"unique"`
}

func (*netrGetDCNameRequest) Opnum() uint16 { return logon.OpnumNetrGetDCName }

// netrGetDCNameResponse carries the [out] parameters and return value of NetrGetDCName.
type netrGetDCNameResponse struct {
	Buffer ndr.WSTR
	Status ndr.DWORD `ndr:"retval"`
}

// NetrGetDCName calls NetrGetDCName (opnum 11) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrGetDCName(rpc ndr.Invoker, serverName ndr.WSTR, domainName *ndr.WSTR) (Buffer ndr.WSTR, err error) {
	req := &netrGetDCNameRequest{
		ServerName: serverName,
		DomainName: domainName,
	}
	var resp netrGetDCNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrGetDCName: %w", err)
		return
	}
	Buffer = resp.Buffer
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrGetDCName failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// lsarGetUserNameRequest is the [in]/[in,out] parameter set of LsarGetUserName: a [unique]
// SystemName string (sent NULL to target the local system), an [in,out] UserName (a double
// pointer, sent NULL), and an [in,out,unique] DomainName (a double pointer, sent NULL).
type lsarGetUserNameRequest struct {
	SystemName *ndr.WSTR                  `ndr:"unique"`
	UserName   *msdtyp.RPC_UNICODE_STRING `ndr:"unique"`
	DomainName *msdtyp.RPC_UNICODE_STRING `ndr:"unique"`
}

func (*lsarGetUserNameRequest) Opnum() uint16 { return lsarpc.OpnumLsarGetUserName }

// lsarGetUserNameResponse is the reply: the [in,out] UserName, the [in,out,unique]
// DomainName, and the NTSTATUS return value.
type lsarGetUserNameResponse struct {
	UserName   *msdtyp.RPC_UNICODE_STRING `ndr:"unique"`
	DomainName *msdtyp.RPC_UNICODE_STRING `ndr:"unique"`
	Status     ndr.DWORD                  `ndr:"retval"`
}

// LsarGetUserName calls LsarGetUserName (opnum 45) to retrieve the name (and domain) of
// the security principal that established the RPC connection ([MS-LSAT] 3.1.4.2).
// SystemName is sent NULL to target the local system; UserName and DomainName are sent
// NULL so the server allocates the returned strings.
func LsarGetUserName(rpc ndr.Invoker) (*msdtyp.RPC_UNICODE_STRING, *msdtyp.RPC_UNICODE_STRING, error) {
	req := &lsarGetUserNameRequest{}
	var resp lsarGetUserNameResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, nil, fmt.Errorf("LsarGetUserName: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.UserName, resp.DomainName, fmt.Errorf("LsarGetUserName failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.UserName, resp.DomainName, nil
}

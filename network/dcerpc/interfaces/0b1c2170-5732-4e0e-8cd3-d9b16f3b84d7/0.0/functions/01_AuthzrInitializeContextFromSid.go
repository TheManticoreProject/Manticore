package functions

import (
	"fmt"

	authzr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/0b1c2170-5732-4e0e-8cd3-d9b16f3b84d7/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	msraa "github.com/TheManticoreProject/Manticore/windows/protocols/ms-raa"
)

// authzrInitializeContextFromSidRequest carries the [in] parameters of AuthzrInitializeContextFromSid.
type authzrInitializeContextFromSidRequest struct {
	Flags           ndr.DWORD
	Sid             msdtyp.RPC_SID
	PExpirationTime *msdtyp.LARGE_INTEGER `ndr:"unique"`
	Identifier      msdtyp.LUID
}

func (*authzrInitializeContextFromSidRequest) Opnum() uint16 {
	return authzr.OpnumAuthzrInitializeContextFromSid
}

// authzrInitializeContextFromSidResponse carries the [out] parameters and return value of AuthzrInitializeContextFromSid.
type authzrInitializeContextFromSidResponse struct {
	ContextHandle msraa.AUTHZR_HANDLE
	Status        ndr.DWORD `ndr:"retval"`
}

// AuthzrInitializeContextFromSid calls AuthzrInitializeContextFromSid (opnum 1) ([MS-RAA] — verify the parameter
// modeling and status handling).
func AuthzrInitializeContextFromSid(rpc ndr.Invoker, flags ndr.DWORD, sid msdtyp.RPC_SID, pExpirationTime *msdtyp.LARGE_INTEGER, identifier msdtyp.LUID) (ContextHandle msraa.AUTHZR_HANDLE, err error) {
	req := &authzrInitializeContextFromSidRequest{
		Flags:           flags,
		Sid:             sid,
		PExpirationTime: pExpirationTime,
		Identifier:      identifier,
	}
	var resp authzrInitializeContextFromSidResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("AuthzrInitializeContextFromSid: %w", err)
		return
	}
	ContextHandle = resp.ContextHandle
	if uint32(resp.Status) != authzr.StatusSuccess {
		err = fmt.Errorf("AuthzrInitializeContextFromSid failed: %s", authzr.StatusString(uint32(resp.Status)))
	}
	return
}

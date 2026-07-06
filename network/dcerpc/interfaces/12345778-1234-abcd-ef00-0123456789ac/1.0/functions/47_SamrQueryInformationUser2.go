package functions

// IDL source: [MS-SAMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/1cd138b9-cc1b-4706-b115-49e53189e32e
// A fetched copy is kept at ms-samr.idl in the interface directory.

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrQueryInformationUser2Request carries the [in] user handle and the information class
// selecting which arm of the returned union is populated.
type samrQueryInformationUser2Request struct {
	UserHandle           mssamr.SAMPR_HANDLE
	UserInformationClass mssamr.USER_INFORMATION_CLASS `ndr:"enum"`
}

func (*samrQueryInformationUser2Request) Opnum() uint16 {
	return samr.OpnumSamrQueryInformationUser2
}

// samrQueryInformationUser2Response is the reply: the [out,switch_is,unique] user info buffer
// (carrying its own discriminant) and the NTSTATUS.
type samrQueryInformationUser2Response struct {
	Buffer *mssamr.SAMPR_USER_INFO_BUFFER `ndr:"unique"`
	Status ndr.DWORD                      `ndr:"retval"`
}

// SamrQueryInformationUser2 calls SamrQueryInformationUser2 (opnum 47), retrieving attributes
// of a user object ([MS-SAMR] 3.1.5.5.5).
func SamrQueryInformationUser2(rpc ndr.Invoker, userHandle mssamr.SAMPR_HANDLE, userInformationClass mssamr.USER_INFORMATION_CLASS) (*mssamr.SAMPR_USER_INFO_BUFFER, error) {
	req := &samrQueryInformationUser2Request{
		UserHandle:           userHandle,
		UserInformationClass: userInformationClass,
	}
	var resp samrQueryInformationUser2Response
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("SamrQueryInformationUser2: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Buffer, fmt.Errorf("SamrQueryInformationUser2 failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Buffer, nil
}

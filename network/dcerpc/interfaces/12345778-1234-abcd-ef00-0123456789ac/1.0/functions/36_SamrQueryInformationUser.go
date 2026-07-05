package functions

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrQueryInformationUserRequest carries the [in] user handle and the information class
// selecting which arm of the returned union is populated.
type samrQueryInformationUserRequest struct {
	UserHandle           mssamr.SAMPR_HANDLE
	UserInformationClass mssamr.USER_INFORMATION_CLASS `ndr:"enum"`
}

func (*samrQueryInformationUserRequest) Opnum() uint16 {
	return samr.OpnumSamrQueryInformationUser
}

// samrQueryInformationUserResponse is the reply: the [out,switch_is,unique] user info buffer
// (carrying its own discriminant) and the NTSTATUS.
type samrQueryInformationUserResponse struct {
	Buffer *mssamr.SAMPR_USER_INFO_BUFFER `ndr:"unique"`
	Status ndr.DWORD                      `ndr:"retval"`
}

// SamrQueryInformationUser calls SamrQueryInformationUser (opnum 36), retrieving attributes
// of a user object ([MS-SAMR] 3.1.5.5.6).
func SamrQueryInformationUser(rpc ndr.Invoker, userHandle mssamr.SAMPR_HANDLE, userInformationClass mssamr.USER_INFORMATION_CLASS) (*mssamr.SAMPR_USER_INFO_BUFFER, error) {
	req := &samrQueryInformationUserRequest{
		UserHandle:           userHandle,
		UserInformationClass: userInformationClass,
	}
	var resp samrQueryInformationUserResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("SamrQueryInformationUser: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Buffer, fmt.Errorf("SamrQueryInformationUser failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Buffer, nil
}

package functions

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// samrQueryInformationUserRequest carries the [in] user handle and the information class
// selecting which arm of the returned union is populated.
type samrQueryInformationUserRequest struct {
	UserHandle           structures.SAMPR_HANDLE
	UserInformationClass structures.USER_INFORMATION_CLASS `ndr:"enum"`
}

func (*samrQueryInformationUserRequest) Opnum() uint16 {
	return samr.OpnumSamrQueryInformationUser
}

// samrQueryInformationUserResponse is the reply: the [out,switch_is,unique] user info buffer
// (carrying its own discriminant) and the NTSTATUS.
type samrQueryInformationUserResponse struct {
	Buffer *structures.SAMPR_USER_INFO_BUFFER `ndr:"unique"`
	Status ndr.DWORD                          `ndr:"retval"`
}

// SamrQueryInformationUser calls SamrQueryInformationUser (opnum 36), retrieving attributes
// of a user object ([MS-SAMR] 3.1.5.5.6).
func SamrQueryInformationUser(rpc ndr.Invoker, userHandle structures.SAMPR_HANDLE, userInformationClass structures.USER_INFORMATION_CLASS) (*structures.SAMPR_USER_INFO_BUFFER, error) {
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

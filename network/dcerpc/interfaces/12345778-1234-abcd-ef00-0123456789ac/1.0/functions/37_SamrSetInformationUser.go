package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrSetInformationUserRequest carries the [in] user handle, the information class, and the
// [in,switch_is] user info buffer (inline, carrying its own discriminant) whose selected arm
// is written to the user object.
type samrSetInformationUserRequest struct {
	UserHandle           mssamr.SAMPR_HANDLE
	UserInformationClass mssamr.USER_INFORMATION_CLASS `ndr:"enum"`
	Buffer               mssamr.SAMPR_USER_INFO_BUFFER
}

func (*samrSetInformationUserRequest) Opnum() uint16 { return samr.OpnumSamrSetInformationUser }

// SamrSetInformationUser calls SamrSetInformationUser (opnum 37), updating attributes of a
// user object ([MS-SAMR] 3.1.5.6.5).
func SamrSetInformationUser(rpc ndr.Invoker, userHandle mssamr.SAMPR_HANDLE, userInformationClass mssamr.USER_INFORMATION_CLASS, buffer mssamr.SAMPR_USER_INFO_BUFFER) error {
	req := &samrSetInformationUserRequest{
		UserHandle:           userHandle,
		UserInformationClass: userInformationClass,
		Buffer:               buffer,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrSetInformationUser: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrSetInformationUser failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}

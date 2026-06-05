package functions

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// samrQueryInformationUser2Request carries the [in] user handle and the information class
// selecting which arm of the returned union is populated.
type samrQueryInformationUser2Request struct {
	UserHandle           structures.SAMPR_HANDLE
	UserInformationClass structures.USER_INFORMATION_CLASS
}

func (*samrQueryInformationUser2Request) Opnum() uint16 {
	return samr.OpnumSamrQueryInformationUser2
}

// samrQueryInformationUser2Response is the reply: the [out,switch_is,unique] user info buffer
// (carrying its own discriminant) and the NTSTATUS.
type samrQueryInformationUser2Response struct {
	Buffer *structures.SAMPR_USER_INFO_BUFFER `ndr:"unique"`
	Status ndr.DWORD                          `ndr:"retval"`
}

// SamrQueryInformationUser2 calls SamrQueryInformationUser2 (opnum 47), retrieving attributes
// of a user object ([MS-SAMR] 3.1.5.5.5).
func SamrQueryInformationUser2(rpc ndr.Invoker, userHandle structures.SAMPR_HANDLE, userInformationClass structures.USER_INFORMATION_CLASS) (*structures.SAMPR_USER_INFO_BUFFER, error) {
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

package functions

import (
	"fmt"

	inetinfo "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82ad4280-036b-11cf-972c-00aa006887b0/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msirp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-irp"
)

// r_IISEnumerateUsersRequest carries the [in] parameters of R_IISEnumerateUsers.
type r_IISEnumerateUsersRequest struct {
	PszServer   *ndr.WSTR `ndr:"unique"`
	DwServiceId ndr.DWORD
	DwInstance  ndr.DWORD
	InfoStruct  msirp.IIS_USER_ENUM_STRUCT
}

func (*r_IISEnumerateUsersRequest) Opnum() uint16 { return inetinfo.OpnumR_IISEnumerateUsers }

// r_IISEnumerateUsersResponse carries the [out] parameters and return value of R_IISEnumerateUsers.
type r_IISEnumerateUsersResponse struct {
	InfoStruct msirp.IIS_USER_ENUM_STRUCT
	Status     ndr.DWORD `ndr:"retval"`
}

// R_IISEnumerateUsers calls R_IISEnumerateUsers (opnum 14) ([MS-IRP] — verify the parameter
// modeling and status handling).
func R_IISEnumerateUsers(rpc ndr.Invoker, pszServer *ndr.WSTR, dwServiceId ndr.DWORD, dwInstance ndr.DWORD, infoStruct msirp.IIS_USER_ENUM_STRUCT) (InfoStruct msirp.IIS_USER_ENUM_STRUCT, err error) {
	req := &r_IISEnumerateUsersRequest{
		PszServer:   pszServer,
		DwServiceId: dwServiceId,
		DwInstance:  dwInstance,
		InfoStruct:  infoStruct,
	}
	var resp r_IISEnumerateUsersResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_IISEnumerateUsers: %w", err)
		return
	}
	InfoStruct = resp.InfoStruct
	if uint32(resp.Status) != inetinfo.ErrorSuccess {
		err = fmt.Errorf("R_IISEnumerateUsers failed: %s", inetinfo.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rOpenSCManagerWRequest carries the [in] parameters of ROpenSCManagerW.
type rOpenSCManagerWRequest struct {
	LpMachineName   *ndr.WSTR `ndr:"unique"`
	LpDatabaseName  *ndr.WSTR `ndr:"unique"`
	DwDesiredAccess ndr.DWORD
}

func (*rOpenSCManagerWRequest) Opnum() uint16 { return svcctl.OpnumROpenSCManagerW }

// rOpenSCManagerWResponse carries the [out] parameters and return value of ROpenSCManagerW.
type rOpenSCManagerWResponse struct {
	LpScHandle structures.LPSC_RPC_HANDLE
	Status     ndr.DWORD `ndr:"retval"`
}

// ROpenSCManagerW calls ROpenSCManagerW (opnum 15) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func ROpenSCManagerW(rpc ndr.Invoker, lpMachineName *ndr.WSTR, lpDatabaseName *ndr.WSTR, dwDesiredAccess ndr.DWORD) (LpScHandle structures.LPSC_RPC_HANDLE, err error) {
	req := &rOpenSCManagerWRequest{
		LpMachineName:   lpMachineName,
		LpDatabaseName:  lpDatabaseName,
		DwDesiredAccess: dwDesiredAccess,
	}
	var resp rOpenSCManagerWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ROpenSCManagerW: %w", err)
		return
	}
	LpScHandle = resp.LpScHandle
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("ROpenSCManagerW failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}

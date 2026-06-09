package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rStartServiceARequest carries the [in] parameters of RStartServiceA.
type rStartServiceARequest struct {
	HService structures.SC_RPC_HANDLE
	Argc     ndr.DWORD
	Argv     []structures.STRING_PTRSA `ndr:"unique,size_is=Argc"`
}

func (*rStartServiceARequest) Opnum() uint16 { return svcctl.OpnumRStartServiceA }

// rStartServiceAResponse carries the [out] parameters and return value of RStartServiceA.
type rStartServiceAResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RStartServiceA calls RStartServiceA (opnum 31) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RStartServiceA(rpc ndr.Invoker, hService structures.SC_RPC_HANDLE, argc ndr.DWORD, argv []structures.STRING_PTRSA) (err error) {
	req := &rStartServiceARequest{
		HService: hService,
		Argc:     argc,
		Argv:     argv,
	}
	var resp rStartServiceAResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RStartServiceA: %w", err)
		return
	}
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RStartServiceA failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}

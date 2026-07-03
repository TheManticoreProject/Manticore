package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnumUnused47Request carries the [in] parameters of OpnumUnused47.
type opnumUnused47Request struct {
}

func (*opnumUnused47Request) Opnum() uint16 { return logon.OpnumOpnumUnused47 }

// opnumUnused47Response carries the [out] parameters and return value of OpnumUnused47.
type opnumUnused47Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// OpnumUnused47 calls OpnumUnused47 (opnum 47) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func OpnumUnused47(rpc ndr.Invoker) (err error) {
	req := &opnumUnused47Request{}
	var resp opnumUnused47Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("OpnumUnused47: %w", err)
		return
	}
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("OpnumUnused47 failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}

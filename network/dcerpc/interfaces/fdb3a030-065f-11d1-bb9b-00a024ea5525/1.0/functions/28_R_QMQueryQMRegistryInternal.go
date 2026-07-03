package functions

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_QMQueryQMRegistryInternalRequest carries the [in] parameters of R_QMQueryQMRegistryInternal.
type r_QMQueryQMRegistryInternalRequest struct {
	DwQueryType ndr.DWORD
}

func (*r_QMQueryQMRegistryInternalRequest) Opnum() uint16 {
	return qmcomm.OpnumR_QMQueryQMRegistryInternal
}

// r_QMQueryQMRegistryInternalResponse carries the [out] parameters and return value of R_QMQueryQMRegistryInternal.
// r_QMQueryQMRegistryInternalResponse carries the [out] parameters and return value of
// R_QMQueryQMRegistryInternal. lplpMQISServer is [out, string] WCHAR**; the outer [out]
// indirection is implicit, leaving an inner [unique] WCHAR* that carries a referent id.
type r_QMQueryQMRegistryInternalResponse struct {
	LplpMQISServer *ndr.WSTR `ndr:"unique"`
	Status         ndr.DWORD `ndr:"retval"`
}

// R_QMQueryQMRegistryInternal calls R_QMQueryQMRegistryInternal (opnum 28) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func R_QMQueryQMRegistryInternal(rpc ndr.Invoker, dwQueryType ndr.DWORD) (LplpMQISServer *ndr.WSTR, err error) {
	req := &r_QMQueryQMRegistryInternalRequest{
		DwQueryType: dwQueryType,
	}
	var resp r_QMQueryQMRegistryInternalResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMQueryQMRegistryInternal: %w", err)
		return
	}
	LplpMQISServer = resp.LplpMQISServer
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("R_QMQueryQMRegistryInternal failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}

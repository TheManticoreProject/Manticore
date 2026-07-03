package functions

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// r_QMCreateObjectInternalRequest carries the [in] parameters of R_QMCreateObjectInternal.
type r_QMCreateObjectInternalRequest struct {
	DwObjectType        ndr.DWORD
	LpwcsPathName       ndr.WSTR
	SDSize              ndr.DWORD
	PSecurityDescriptor []uint8 `ndr:"unique,size_is=SDSize"`
	Cp                  ndr.DWORD
	AProp               []ndr.DWORD          `ndr:"ref,size_is=Cp"`
	ApVar               []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
}

func (*r_QMCreateObjectInternalRequest) Opnum() uint16 { return qmcomm.OpnumR_QMCreateObjectInternal }

// r_QMCreateObjectInternalResponse carries the [out] parameters and return value of R_QMCreateObjectInternal.
type r_QMCreateObjectInternalResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_QMCreateObjectInternal calls R_QMCreateObjectInternal (opnum 6) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func R_QMCreateObjectInternal(rpc ndr.Invoker, dwObjectType ndr.DWORD, lpwcsPathName ndr.WSTR, sDSize ndr.DWORD, pSecurityDescriptor []uint8, cp ndr.DWORD, aProp []ndr.DWORD, apVar []msmqmq.PROPVARIANT) (err error) {
	req := &r_QMCreateObjectInternalRequest{
		DwObjectType:        dwObjectType,
		LpwcsPathName:       lpwcsPathName,
		SDSize:              sDSize,
		PSecurityDescriptor: pSecurityDescriptor,
		Cp:                  cp,
		AProp:               aProp,
		ApVar:               apVar,
	}
	var resp r_QMCreateObjectInternalResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMCreateObjectInternal: %w", err)
		return
	}
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("R_QMCreateObjectInternal failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}

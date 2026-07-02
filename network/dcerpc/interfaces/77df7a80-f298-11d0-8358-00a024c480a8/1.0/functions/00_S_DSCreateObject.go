package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// s_DSCreateObjectRequest carries the [in] parameters of S_DSCreateObject.
type s_DSCreateObjectRequest struct {
	DwObjectType       ndr.DWORD
	PwcsPathName       *ndr.WSTR `ndr:"unique"`
	DwSDLength         ndr.DWORD
	SecurityDescriptor []uint8 `ndr:"ref,size_is=DwSDLength"`
	Cp                 ndr.DWORD
	AProp              []ndr.DWORD          `ndr:"ref,size_is=Cp"`
	ApVar              []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
	PObjGuid           *dtyp.GUID           `ndr:"unique"`
}

func (*s_DSCreateObjectRequest) Opnum() uint16 { return dscomm.OpnumS_DSCreateObject }

// s_DSCreateObjectResponse carries the [out] parameters and return value of S_DSCreateObject.
type s_DSCreateObjectResponse struct {
	PObjGuid *dtyp.GUID `ndr:"unique"`
	Status   ndr.DWORD  `ndr:"retval"`
}

// S_DSCreateObject calls S_DSCreateObject (opnum 0) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSCreateObject(rpc ndr.Invoker, dwObjectType ndr.DWORD, pwcsPathName *ndr.WSTR, dwSDLength ndr.DWORD, securityDescriptor []uint8, cp ndr.DWORD, aProp []ndr.DWORD, apVar []msmqmq.PROPVARIANT, pObjGuid *dtyp.GUID) (PObjGuid *dtyp.GUID, err error) {
	req := &s_DSCreateObjectRequest{
		DwObjectType:       dwObjectType,
		PwcsPathName:       pwcsPathName,
		DwSDLength:         dwSDLength,
		SecurityDescriptor: securityDescriptor,
		Cp:                 cp,
		AProp:              aProp,
		ApVar:              apVar,
		PObjGuid:           pObjGuid,
	}
	var resp s_DSCreateObjectResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSCreateObject: %w", err)
		return
	}
	PObjGuid = resp.PObjGuid
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSCreateObject failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}

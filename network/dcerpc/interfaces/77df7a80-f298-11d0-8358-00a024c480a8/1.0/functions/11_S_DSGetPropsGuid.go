package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqds "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqds"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// s_DSGetPropsGuidRequest carries the [in] parameters of S_DSGetPropsGuid.
type s_DSGetPropsGuidRequest struct {
	DwObjectType           ndr.DWORD
	PGuid                  *dtyp.GUID `ndr:"unique"`
	Cp                     ndr.DWORD
	AProp                  []ndr.DWORD          `ndr:"ref,size_is=Cp"`
	ApVar                  []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
	PhServerAuth           msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
}

func (*s_DSGetPropsGuidRequest) Opnum() uint16 { return dscomm.OpnumS_DSGetPropsGuid }

// s_DSGetPropsGuidResponse carries the [out] parameters and return value of S_DSGetPropsGuid.
type s_DSGetPropsGuidResponse struct {
	ApVar                  []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
	PbServerSignature      []uint8              `ndr:"ref,conformant"`
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
	Status                 ndr.DWORD `ndr:"retval"`
}

// S_DSGetPropsGuid calls S_DSGetPropsGuid (opnum 11) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSGetPropsGuid(rpc ndr.Invoker, dwObjectType ndr.DWORD, pGuid *dtyp.GUID, cp ndr.DWORD, aProp []ndr.DWORD, apVar []msmqmq.PROPVARIANT, phServerAuth msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE, pdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE) (ApVar []msmqmq.PROPVARIANT, PbServerSignature []uint8, PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE, err error) {
	req := &s_DSGetPropsGuidRequest{
		DwObjectType:           dwObjectType,
		PGuid:                  pGuid,
		Cp:                     cp,
		AProp:                  aProp,
		ApVar:                  apVar,
		PhServerAuth:           phServerAuth,
		PdwServerSignatureSize: pdwServerSignatureSize,
	}
	var resp s_DSGetPropsGuidResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSGetPropsGuid: %w", err)
		return
	}
	ApVar = resp.ApVar
	PbServerSignature = resp.PbServerSignature
	PdwServerSignatureSize = resp.PdwServerSignatureSize
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSGetPropsGuid failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}

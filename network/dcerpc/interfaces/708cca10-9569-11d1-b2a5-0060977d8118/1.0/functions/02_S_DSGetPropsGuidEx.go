package functions

import (
	"fmt"

	dscomm2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/708cca10-9569-11d1-b2a5-0060977d8118/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	msmqds "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqds"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// s_DSGetPropsGuidExRequest carries the [in] parameters of S_DSGetPropsGuidEx.
type s_DSGetPropsGuidExRequest struct {
	DwObjectType           ndr.DWORD
	PGuid                  *msdtyp.GUID `ndr:"unique"`
	Cp                     ndr.DWORD
	AProp                  []msmqmq.PROPID      `ndr:"ref,size_is=Cp"`
	ApVar                  []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
	PhServerAuth           msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
}

func (*s_DSGetPropsGuidExRequest) Opnum() uint16 { return dscomm2.OpnumS_DSGetPropsGuidEx }

// s_DSGetPropsGuidExResponse carries the [out] parameters and return value of S_DSGetPropsGuidEx.
type s_DSGetPropsGuidExResponse struct {
	ApVar                  []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
	PbServerSignature      []uint8              `ndr:"ref,conformant"`
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
	Status                 ndr.DWORD `ndr:"retval"`
}

// S_DSGetPropsGuidEx calls S_DSGetPropsGuidEx (opnum 2) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSGetPropsGuidEx(rpc ndr.Invoker, dwObjectType ndr.DWORD, pGuid *msdtyp.GUID, cp ndr.DWORD, aProp []msmqmq.PROPID, apVar []msmqmq.PROPVARIANT, phServerAuth msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE, pdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE) (ApVar []msmqmq.PROPVARIANT, PbServerSignature []uint8, PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE, err error) {
	req := &s_DSGetPropsGuidExRequest{
		DwObjectType:           dwObjectType,
		PGuid:                  pGuid,
		Cp:                     cp,
		AProp:                  aProp,
		ApVar:                  apVar,
		PhServerAuth:           phServerAuth,
		PdwServerSignatureSize: pdwServerSignatureSize,
	}
	var resp s_DSGetPropsGuidExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSGetPropsGuidEx: %w", err)
		return
	}
	ApVar = resp.ApVar
	PbServerSignature = resp.PbServerSignature
	PdwServerSignatureSize = resp.PdwServerSignatureSize
	if uint32(resp.Status) != dscomm2.StatusSuccess {
		err = fmt.Errorf("S_DSGetPropsGuidEx failed: %s", dscomm2.StatusString(uint32(resp.Status)))
	}
	return
}

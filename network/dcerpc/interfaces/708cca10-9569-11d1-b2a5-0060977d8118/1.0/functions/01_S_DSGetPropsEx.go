package functions

import (
	"fmt"

	dscomm2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/708cca10-9569-11d1-b2a5-0060977d8118/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqds "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqds"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// s_DSGetPropsExRequest carries the [in] parameters of S_DSGetPropsEx.
type s_DSGetPropsExRequest struct {
	DwObjectType           ndr.DWORD
	PwcsPathName           ndr.WSTR
	Cp                     ndr.DWORD
	AProp                  []msmqmq.PROPID      `ndr:"ref,size_is=Cp"`
	ApVar                  []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
	PhServerAuth           msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
}

func (*s_DSGetPropsExRequest) Opnum() uint16 { return dscomm2.OpnumS_DSGetPropsEx }

// s_DSGetPropsExResponse carries the [out] parameters and return value of S_DSGetPropsEx.
type s_DSGetPropsExResponse struct {
	ApVar                  []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
	PbServerSignature      []uint8              `ndr:"ref,conformant"`
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
	Status                 ndr.DWORD `ndr:"retval"`
}

// S_DSGetPropsEx calls S_DSGetPropsEx (opnum 1) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSGetPropsEx(rpc ndr.Invoker, dwObjectType ndr.DWORD, pwcsPathName ndr.WSTR, cp ndr.DWORD, aProp []msmqmq.PROPID, apVar []msmqmq.PROPVARIANT, phServerAuth msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE, pdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE) (ApVar []msmqmq.PROPVARIANT, PbServerSignature []uint8, PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE, err error) {
	req := &s_DSGetPropsExRequest{
		DwObjectType:           dwObjectType,
		PwcsPathName:           pwcsPathName,
		Cp:                     cp,
		AProp:                  aProp,
		ApVar:                  apVar,
		PhServerAuth:           phServerAuth,
		PdwServerSignatureSize: pdwServerSignatureSize,
	}
	var resp s_DSGetPropsExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSGetPropsEx: %w", err)
		return
	}
	ApVar = resp.ApVar
	PbServerSignature = resp.PbServerSignature
	PdwServerSignatureSize = resp.PdwServerSignatureSize
	if uint32(resp.Status) != dscomm2.StatusSuccess {
		err = fmt.Errorf("S_DSGetPropsEx failed: %s", dscomm2.StatusString(uint32(resp.Status)))
	}
	return
}

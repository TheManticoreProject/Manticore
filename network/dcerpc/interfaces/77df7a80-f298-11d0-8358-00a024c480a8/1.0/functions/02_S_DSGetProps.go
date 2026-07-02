package functions

import (
	"fmt"

	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqds "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqds"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// s_DSGetPropsRequest carries the [in] parameters of S_DSGetProps.
type s_DSGetPropsRequest struct {
	DwObjectType           ndr.DWORD
	PwcsPathName           ndr.WSTR
	Cp                     ndr.DWORD
	AProp                  []ndr.DWORD          `ndr:"ref,size_is=Cp"`
	ApVar                  []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
	PhServerAuth           msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
}

func (*s_DSGetPropsRequest) Opnum() uint16 { return dscomm.OpnumS_DSGetProps }

// s_DSGetPropsResponse carries the [out] parameters and return value of S_DSGetProps.
type s_DSGetPropsResponse struct {
	ApVar                  []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
	PbServerSignature      []uint8              `ndr:"ref,conformant"`
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
	Status                 ndr.DWORD `ndr:"retval"`
}

// S_DSGetProps calls S_DSGetProps (opnum 2) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSGetProps(rpc ndr.Invoker, dwObjectType ndr.DWORD, pwcsPathName ndr.WSTR, cp ndr.DWORD, aProp []ndr.DWORD, apVar []msmqmq.PROPVARIANT, phServerAuth msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE, pdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE) (ApVar []msmqmq.PROPVARIANT, PbServerSignature []uint8, PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE, err error) {
	req := &s_DSGetPropsRequest{
		DwObjectType:           dwObjectType,
		PwcsPathName:           pwcsPathName,
		Cp:                     cp,
		AProp:                  aProp,
		ApVar:                  apVar,
		PhServerAuth:           phServerAuth,
		PdwServerSignatureSize: pdwServerSignatureSize,
	}
	var resp s_DSGetPropsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSGetProps: %w", err)
		return
	}
	ApVar = resp.ApVar
	PbServerSignature = resp.PbServerSignature
	PdwServerSignatureSize = resp.PdwServerSignatureSize
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSGetProps failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqds "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqds"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// s_DSLookupNextRequest carries the [in] parameters of S_DSLookupNext.
type s_DSLookupNextRequest struct {
	Handle                 msmqds.PCONTEXT_HANDLE_TYPE
	DwSize                 msmqds.LPBOUNDED_PROPERTIES
	PhServerAuth           msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
}

func (*s_DSLookupNextRequest) Opnum() uint16 { return dscomm.OpnumS_DSLookupNext }

// s_DSLookupNextResponse carries the [out] parameters and return value of S_DSLookupNext.
type s_DSLookupNextResponse struct {
	DwOutSize              ndr.DWORD
	PbBuffer               []msmqmq.PROPVARIANT `ndr:"ref,conformant,varying,length_is=DwOutSize"`
	PbServerSignature      []uint8              `ndr:"ref,conformant"`
	PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE
	Status                 ndr.DWORD `ndr:"retval"`
}

// S_DSLookupNext calls S_DSLookupNext (opnum 7) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSLookupNext(rpc ndr.Invoker, handle msmqds.PCONTEXT_HANDLE_TYPE, dwSize msmqds.LPBOUNDED_PROPERTIES, phServerAuth msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE, pdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE) (DwOutSize ndr.DWORD, PbBuffer []msmqmq.PROPVARIANT, PbServerSignature []uint8, PdwServerSignatureSize msmqds.LPBOUNDED_SIGNATURE_SIZE, err error) {
	req := &s_DSLookupNextRequest{
		Handle:                 handle,
		DwSize:                 dwSize,
		PhServerAuth:           phServerAuth,
		PdwServerSignatureSize: pdwServerSignatureSize,
	}
	var resp s_DSLookupNextResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSLookupNext: %w", err)
		return
	}
	DwOutSize = resp.DwOutSize
	PbBuffer = resp.PbBuffer
	PbServerSignature = resp.PbServerSignature
	PdwServerSignatureSize = resp.PdwServerSignatureSize
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSLookupNext failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	FrsTransport "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/897e2e5f-93f3-4376-9c9c-fd2277495c27/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfrs2 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs2"
)

// rdcGetSignaturesRequest carries the [in] parameters of RdcGetSignatures.
type rdcGetSignaturesRequest struct {
	ServerContext msfrs2.PFRS_SERVER_CONTEXT
	Level         uint8
	Offset        uint64
	Length        ndr.DWORD
}

func (*rdcGetSignaturesRequest) Opnum() uint16 { return FrsTransport.OpnumRdcGetSignatures }

// rdcGetSignaturesResponse carries the [out] parameters and return value of RdcGetSignatures.
type rdcGetSignaturesResponse struct {
	Buffer   []uint8 `ndr:"ref,size_is=Length,varying"`
	SizeRead ndr.DWORD
	Status   ndr.DWORD `ndr:"retval"`
}

// RdcGetSignatures calls RdcGetSignatures (opnum 9) ([MS-FRS2] — verify the parameter
// modeling and status handling).
func RdcGetSignatures(rpc ndr.Invoker, serverContext msfrs2.PFRS_SERVER_CONTEXT, level uint8, offset uint64, length ndr.DWORD) (Buffer []uint8, SizeRead ndr.DWORD, err error) {
	req := &rdcGetSignaturesRequest{
		ServerContext: serverContext,
		Level:         level,
		Offset:        offset,
		Length:        length,
	}
	var resp rdcGetSignaturesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RdcGetSignatures: %w", err)
		return
	}
	Buffer = resp.Buffer
	SizeRead = resp.SizeRead
	if uint32(resp.Status) != FrsTransport.StatusSuccess {
		err = fmt.Errorf("RdcGetSignatures failed: %s", FrsTransport.StatusString(uint32(resp.Status)))
	}
	return
}

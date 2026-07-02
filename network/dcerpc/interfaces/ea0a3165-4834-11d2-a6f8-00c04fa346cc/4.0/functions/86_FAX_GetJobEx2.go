package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_GetJobEx2Request carries the [in] parameters of FAX_GetJobEx2.
type fAX_GetJobEx2Request struct {
	DwlMessageID uint64
	Level        ndr.DWORD
}

func (*fAX_GetJobEx2Request) Opnum() uint16 { return fax.OpnumFAX_GetJobEx2 }

// fAX_GetJobEx2Response carries the [out] parameters and return value of FAX_GetJobEx2.
type fAX_GetJobEx2Response struct {
	Buffer     []byte `ndr:"unique,conformant"`
	BufferSize ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// FAX_GetJobEx2 calls FAX_GetJobEx2 (opnum 86) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetJobEx2(rpc ndr.Invoker, dwlMessageID uint64, level ndr.DWORD) (Buffer []byte, BufferSize ndr.DWORD, err error) {
	req := &fAX_GetJobEx2Request{
		DwlMessageID: dwlMessageID,
		Level:        level,
	}
	var resp fAX_GetJobEx2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetJobEx2: %w", err)
		return
	}
	Buffer = resp.Buffer
	BufferSize = resp.BufferSize
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetJobEx2 failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}

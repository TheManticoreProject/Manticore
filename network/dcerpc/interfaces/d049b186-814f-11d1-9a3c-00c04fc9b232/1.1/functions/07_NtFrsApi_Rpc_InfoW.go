package functions

import (
	"fmt"

	frsapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/d049b186-814f-11d1-9a3c-00c04fc9b232/1.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// ntFrsApi_Rpc_InfoWRequest carries the [in] parameters of NtFrsApi_Rpc_InfoW.
type ntFrsApi_Rpc_InfoWRequest struct {
	BlobSize ndr.DWORD
	Blob     []uint8 `ndr:"unique,size_is=BlobSize"`
}

func (*ntFrsApi_Rpc_InfoWRequest) Opnum() uint16 { return frsapi.OpnumNtFrsApi_Rpc_InfoW }

// ntFrsApi_Rpc_InfoWResponse carries the [out] parameters and return value of NtFrsApi_Rpc_InfoW.
type ntFrsApi_Rpc_InfoWResponse struct {
	// Blob is the [in, out] buffer returned by the server; its maximum_count is read
	// from the wire, so no size_is sibling is needed on the response.
	Blob   []uint8   `ndr:"unique"`
	Status ndr.DWORD `ndr:"retval"`
}

// NtFrsApi_Rpc_InfoW calls NtFrsApi_Rpc_InfoW (opnum 7) ([MS-FRS1] section 3.2.4.3).
func NtFrsApi_Rpc_InfoW(rpc ndr.Invoker, blobSize ndr.DWORD, blob []uint8) (Blob []uint8, err error) {
	req := &ntFrsApi_Rpc_InfoWRequest{
		BlobSize: blobSize,
		Blob:     blob,
	}
	var resp ntFrsApi_Rpc_InfoWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NtFrsApi_Rpc_InfoW: %w", err)
		return
	}
	Blob = resp.Blob
	if uint32(resp.Status) != frsapi.StatusSuccess {
		err = fmt.Errorf("NtFrsApi_Rpc_InfoW failed: %s", frsapi.StatusString(uint32(resp.Status)))
	}
	return
}

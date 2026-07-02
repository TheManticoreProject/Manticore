package functions

import (
	"fmt"

	FileServerVssAgent "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/a8e0653c-2744-4389-a61d-7373df8b2292/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// getSupportedVersionRequest carries the [in] parameters of GetSupportedVersion.
type getSupportedVersionRequest struct {
}

func (*getSupportedVersionRequest) Opnum() uint16 { return FileServerVssAgent.OpnumGetSupportedVersion }

// getSupportedVersionResponse carries the [out] parameters and return value of GetSupportedVersion.
type getSupportedVersionResponse struct {
	MinVersion ndr.DWORD
	MaxVersion ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// GetSupportedVersion calls GetSupportedVersion (opnum 0) ([MS-FSRVP] — verify the parameter
// modeling and status handling).
func GetSupportedVersion(rpc ndr.Invoker) (MinVersion ndr.DWORD, MaxVersion ndr.DWORD, err error) {
	req := &getSupportedVersionRequest{}
	var resp getSupportedVersionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("GetSupportedVersion: %w", err)
		return
	}
	MinVersion = resp.MinVersion
	MaxVersion = resp.MaxVersion
	if uint32(resp.Status) != FileServerVssAgent.StatusSuccess {
		err = fmt.Errorf("GetSupportedVersion failed: %s", FileServerVssAgent.StatusString(uint32(resp.Status)))
	}
	return
}

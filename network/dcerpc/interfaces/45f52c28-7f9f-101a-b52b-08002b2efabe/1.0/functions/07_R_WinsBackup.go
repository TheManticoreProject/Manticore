package functions

// IDL source: [MS-RAIW] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-raiw/e59461f5-5486-4ec3-9ad6-14b784c1ecd6
// A fetched copy is kept at ms-raiw.idl in the interface directory.

import (
	"fmt"

	winsif "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/45f52c28-7f9f-101a-b52b-08002b2efabe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_WinsBackupRequest carries the [in] parameters of R_WinsBackup.
type r_WinsBackupRequest struct {
	PBackupPath  ndr.STR
	FIncremental int16
}

func (*r_WinsBackupRequest) Opnum() uint16 { return winsif.OpnumR_WinsBackup }

// r_WinsBackupResponse carries the [out] parameters and return value of R_WinsBackup.
type r_WinsBackupResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsBackup calls R_WinsBackup (opnum 7) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsBackup(rpc ndr.Invoker, pBackupPath ndr.STR, fIncremental int16) (err error) {
	req := &r_WinsBackupRequest{
		PBackupPath:  pBackupPath,
		FIncremental: fIncremental,
	}
	var resp r_WinsBackupResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsBackup: %w", err)
		return
	}
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsBackup failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}

package functions

import (
	"fmt"

	inetinfo "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82ad4280-036b-11cf-972c-00aa006887b0/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_FtpClearStatistics2Request carries the [in] parameters of R_FtpClearStatistics2.
type r_FtpClearStatistics2Request struct {
	PszServer  *ndr.WSTR `ndr:"unique"`
	DwInstance ndr.DWORD
}

func (*r_FtpClearStatistics2Request) Opnum() uint16 { return inetinfo.OpnumR_FtpClearStatistics2 }

// r_FtpClearStatistics2Response carries the [out] parameters and return value of R_FtpClearStatistics2.
type r_FtpClearStatistics2Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_FtpClearStatistics2 calls R_FtpClearStatistics2 (opnum 13) ([MS-IRP] — verify the parameter
// modeling and status handling).
func R_FtpClearStatistics2(rpc ndr.Invoker, pszServer *ndr.WSTR, dwInstance ndr.DWORD) (err error) {
	req := &r_FtpClearStatistics2Request{
		PszServer:  pszServer,
		DwInstance: dwInstance,
	}
	var resp r_FtpClearStatistics2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_FtpClearStatistics2: %w", err)
		return
	}
	if uint32(resp.Status) != inetinfo.ErrorSuccess {
		err = fmt.Errorf("R_FtpClearStatistics2 failed: %s", inetinfo.StatusString(uint32(resp.Status)))
	}
	return
}

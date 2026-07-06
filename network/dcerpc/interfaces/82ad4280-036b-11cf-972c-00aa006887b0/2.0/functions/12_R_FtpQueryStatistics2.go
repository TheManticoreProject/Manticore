package functions

// IDL source: [MS-IRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-irp/ed7e5940-9700-4a1f-8555-de29f99fe115
// A fetched copy is kept at ms-irp.idl in the interface directory.

import (
	"fmt"

	inetinfo "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82ad4280-036b-11cf-972c-00aa006887b0/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msirp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-irp"
)

// r_FtpQueryStatistics2Request carries the [in] parameters of R_FtpQueryStatistics2.
type r_FtpQueryStatistics2Request struct {
	PszServer  *ndr.WSTR `ndr:"unique"`
	DwLevel    ndr.DWORD
	DwInstance ndr.DWORD
	DwReserved ndr.DWORD
}

func (*r_FtpQueryStatistics2Request) Opnum() uint16 { return inetinfo.OpnumR_FtpQueryStatistics2 }

// r_FtpQueryStatistics2Response carries the [out] parameters and return value of R_FtpQueryStatistics2.
type r_FtpQueryStatistics2Response struct {
	InfoStruct msirp.FTP_STATISTICS_STRUCT
	Status     ndr.DWORD `ndr:"retval"`
}

// R_FtpQueryStatistics2 calls R_FtpQueryStatistics2 (opnum 12) ([MS-IRP] — verify the parameter
// modeling and status handling).
func R_FtpQueryStatistics2(rpc ndr.Invoker, pszServer *ndr.WSTR, dwLevel ndr.DWORD, dwInstance ndr.DWORD, dwReserved ndr.DWORD) (InfoStruct msirp.FTP_STATISTICS_STRUCT, err error) {
	req := &r_FtpQueryStatistics2Request{
		PszServer:  pszServer,
		DwLevel:    dwLevel,
		DwInstance: dwInstance,
		DwReserved: dwReserved,
	}
	var resp r_FtpQueryStatistics2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_FtpQueryStatistics2: %w", err)
		return
	}
	InfoStruct = resp.InfoStruct
	if uint32(resp.Status) != inetinfo.ErrorSuccess {
		err = fmt.Errorf("R_FtpQueryStatistics2 failed: %s", inetinfo.StatusString(uint32(resp.Status)))
	}
	return
}
